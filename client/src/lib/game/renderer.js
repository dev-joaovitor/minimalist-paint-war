// Canvas renderer. Pure functions called every frame with the latest snapshot;
// kept out of Svelte reactivity for performance. Fixed whole-map camera.

import { colors, teamColor } from '../stores/game.svelte.js';

const FACE_DIR = [
	[1, 0],
	[0.7071, 0.7071],
	[0, 1],
	[-0.7071, 0.7071],
	[-1, 0],
	[-0.7071, -0.7071],
	[0, -1],
	[0.7071, -0.7071]
];
const PLAYER_HALF = 13;
const FLAG_HIT_R = 16; // capture radius; mirrors FlagHitR on the server

/**
 * @param {CanvasRenderingContext2D} ctx
 * @param {any} map
 * @param {any} snapshot
 * @param {string} meId
 */
export function draw(ctx, map, snapshot, meId) {
	const W = map.width;
	const H = map.height;

	ctx.fillStyle = colors.bg;
	ctx.fillRect(0, 0, W, H);

	drawBase(ctx, map.redBase, colors.RED);
	drawBase(ctx, map.greenBase, colors.GREEN);

	map.obstacles.forEach((/** @type {any} */ o, /** @type {number} */ i) => {
		ctx.fillStyle = i % 2 === 0 ? colors.obstacleA : colors.obstacleB;
		ctx.fillRect(o.x, o.y, o.w, o.h);
	});

	/** @type {Record<number, string>} */
	const flagTeam = {};
	if (snapshot && snapshot.flags) {
		snapshot.flags.forEach((/** @type {any} */ f) => (flagTeam[f.id] = f.team));
	}
	map.flags.forEach((/** @type {any} */ f) =>
		drawFlag(ctx, f.x, f.y, teamColor(flagTeam[f.id] || ''))
	);

	if (!snapshot) return;

	snapshot.projectiles.forEach((/** @type {any} */ p) => {
		ctx.fillStyle = teamColor(p.team);
		ctx.beginPath();
		ctx.arc(p.x, p.y, 5, 0, Math.PI * 2);
		ctx.fill();
	});

	snapshot.players.forEach((/** @type {any} */ p) => drawPlayer(ctx, p, meId));
}

/**
 * @param {CanvasRenderingContext2D} ctx
 * @param {any} base
 * @param {string} color
 */
function drawBase(ctx, base, color) {
	if (!base) return;
	ctx.save();
	ctx.globalAlpha = 0.12;
	ctx.fillStyle = color;
	ctx.fillRect(base.x, base.y, base.w, base.h);
	ctx.restore();
	ctx.strokeStyle = color;
	ctx.lineWidth = 6;
	ctx.strokeRect(base.x + 3, base.y + 3, base.w - 6, base.h - 6);
}

/**
 * @param {CanvasRenderingContext2D} ctx
 * @param {number} x
 * @param {number} y
 * @param {string} color
 */
function drawFlag(ctx, x, y, color) {
	ctx.strokeStyle = '#cfcfcf';
	ctx.lineWidth = 3;
	ctx.beginPath();
	ctx.moveTo(x, y - 22);
	ctx.lineTo(x, y + 14);
	ctx.stroke();

	ctx.fillStyle = color;
	ctx.beginPath();
	ctx.moveTo(x, y - 22);
	ctx.lineTo(x + 22, y - 14);
	ctx.lineTo(x, y - 6);
	ctx.closePath();
	ctx.fill();
}

/**
 * @param {CanvasRenderingContext2D} ctx
 * @param {any} p
 * @param {string} meId
 */
function drawPlayer(ctx, p, meId) {
	ctx.save();
	if (p.dead) ctx.globalAlpha = 0.25;

	ctx.fillStyle = teamColor(p.team);
	ctx.fillRect(p.x - PLAYER_HALF, p.y - PLAYER_HALF, PLAYER_HALF * 2, PLAYER_HALF * 2);

	if (p.id === meId) {
		ctx.strokeStyle = '#fff';
		ctx.lineWidth = 2;
		ctx.strokeRect(p.x - PLAYER_HALF, p.y - PLAYER_HALF, PLAYER_HALF * 2, PLAYER_HALF * 2);
	}

	if (p.invuln && !p.dead) {
		ctx.strokeStyle = 'rgba(255, 255, 255, 0.85)';
		ctx.lineWidth = 2;
		ctx.beginPath();
		ctx.arc(p.x, p.y, PLAYER_HALF + 7, 0, Math.PI * 2);
		ctx.stroke();
	}

	if (!p.dead) {
		const d = FACE_DIR[p.face] || FACE_DIR[0];
		ctx.strokeStyle = '#111';
		ctx.lineWidth = 7;
		ctx.beginPath();
		ctx.moveTo(p.x, p.y);
		ctx.lineTo(p.x + d[0] * 18, p.y + d[1] * 18);
		ctx.stroke();
	}
	ctx.restore();
}
