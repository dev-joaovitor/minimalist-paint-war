// Central reactive game state. Server messages mutate this object; components
// and the render loop read from it.

export const colors = {
	bg: '#404040',
	RED: '#E74C3C',
	GREEN: '#7ED321',
	neutral: '#E0E0E0',
	obstacleA: '#F1C40F',
	obstacleB: '#E67E22'
};

/** @param {string} team */
export function teamColor(team) {
	if (team === 'RED') return colors.RED;
	if (team === 'GREEN') return colors.GREEN;
	return colors.neutral;
}

const emptyMe = { id: '', username: '', role: '', team: '' };

/**
 * @typedef {{ code: string, message: string }} GameError
 * @typedef {Object} GameState
 * @property {'disconnected'|'connecting'|'connected'} connection
 * @property {GameError|null} error
 * @property {{ id: string, username: string, role: string, team: string }} me
 * @property {string} phase
 * @property {boolean} isLeader
 * @property {{ leaderId: string, players: any[], leaderboard: any[] }} lobby
 * @property {any} map
 * @property {number} endsAt
 * @property {any} snapshot
 * @property {any} scoreboard
 */

/** @type {GameState} */
export const game = $state({
	connection: 'disconnected', // 'disconnected' | 'connecting' | 'connected'
	error: null, // { code, message }
	me: { ...emptyMe },
	phase: 'LOBBY', // LOBBY | PLAYING | SPECTATOR | SCOREBOARD
	isLeader: false,
	lobby: { leaderId: '', players: [], leaderboard: [] },
	map: null,
	endsAt: 0,
	snapshot: null,
	scoreboard: null
});

export function resetMatch() {
	game.map = null;
	game.snapshot = null;
	game.scoreboard = null;
	game.endsAt = 0;
}

export function resetSession() {
	game.me = { ...emptyMe };
	game.isLeader = false;
	game.phase = 'LOBBY';
	resetMatch();
}
