// WebSocket client: single connection, dispatch to the reactive store, and
// exponential-backoff reconnect that re-sends join.

import { game, resetMatch, resetSession } from '../stores/game.svelte.js';

/** @typedef {import('../stores/game.svelte.js').GameError} GameError */

// Resolve the WebSocket endpoint: an explicit VITE_WS_URL wins; otherwise derive
// it from the page origin so a single-origin (reverse-proxied) cloud deploy works
// without a rebuild. Falls back to the local dev server during SSR.
function resolveWsUrl() {
	if (import.meta.env.VITE_WS_URL) return import.meta.env.VITE_WS_URL;
	if (typeof window !== 'undefined') {
		const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
		return `${proto}://${window.location.host}/ws`;
	}
	return 'ws://localhost:8080/ws';
}

const WS_URL = resolveWsUrl();

/** @type {WebSocket | null} */
let ws = null;
let username = '';
let manualClose = false;
let reconnectDelay = 500;
/** @type {ReturnType<typeof setTimeout> | undefined} */
let reconnectTimer;

/** @param {string} name */
export function connect(name) {
	username = name;
	manualClose = false;
	game.error = null;
	open();
}

export function disconnect() {
	manualClose = true;
	username = '';
	clearTimeout(reconnectTimer);
	if (ws) ws.close();
	resetSession();
	game.connection = 'disconnected';
}

/** @param {string} type @param {any} [data] */
export function send(type, data) {
	if (ws && ws.readyState === WebSocket.OPEN) {
		ws.send(JSON.stringify({ type, data: data ?? {} }));
	}
}

function open() {
	game.connection = 'connecting';
	ws = new WebSocket(WS_URL);
	ws.onopen = () => {
		game.connection = 'connected';
		reconnectDelay = 500;
		send('join', { username });
	};
	ws.onmessage = (ev) => {
		try {
			dispatch(JSON.parse(ev.data));
		} catch {
			/* ignore malformed frames */
		}
	};
	ws.onclose = () => {
		game.connection = 'disconnected';
		if (!manualClose && username) scheduleReconnect();
	};
	ws.onerror = () => {
		/* the close handler drives reconnection */
	};
}

function scheduleReconnect() {
	clearTimeout(reconnectTimer);
	reconnectTimer = setTimeout(open, reconnectDelay);
	reconnectDelay = Math.min(reconnectDelay * 2, 5000);
}

/** @param {{ type: string, data: any }} msg */
function dispatch(msg) {
	const { type, data } = msg;
	switch (type) {
		case 'joined':
			game.me = {
				id: data.userId,
				username: data.username,
				role: data.role,
				team: data.team
			};
			game.error = null;
			break;

		case 'error':
			handleError(data);
			break;

		case 'you_are_leader':
			game.isLeader = true;
			break;

		case 'you_are_spectator':
			game.me.role = 'SPECTATOR';
			break;

		case 'lobby_state':
			game.lobby = data;
			game.phase = 'LOBBY';
			game.isLeader = data.leaderId === game.me.id;
			resetMatch();
			{
				const row = data.players.find((/** @type {any} */ p) => p.id === game.me.id);
				if (row) {
					game.me.team = row.team;
					game.me.role = row.role;
				}
			}
			break;

		case 'match_start':
			game.map = data.map;
			game.endsAt = data.endsAt;
			game.scoreboard = null;
			game.phase = game.me.role === 'SPECTATOR' ? 'SPECTATOR' : 'PLAYING';
			break;

		case 'state':
			game.snapshot = data;
			break;

		case 'scoreboard':
			game.scoreboard = data;
			game.phase = 'SCOREBOARD';
			break;

		case 'pong':
			break;
	}
}

/** @param {GameError} data */
function handleError(data) {
	game.error = data;
	// Fatal join errors: stop reconnecting and return to the login screen.
	if (data.code === 'USERNAME_IN_USE' || data.code === 'INVALID_USERNAME') {
		manualClose = true;
		username = '';
		if (ws) ws.close();
		resetSession();
		game.connection = 'disconnected';
	}
}
