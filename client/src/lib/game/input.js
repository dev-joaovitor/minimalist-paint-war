// Keyboard input: WASD / arrows for 8-directional movement, space to shoot.
// Edge-triggered — sends an input message only when the intent changes (the
// server keeps the last input until the next one arrives).

let dx = 0;
let dy = 0;
let shoot = false;
let prev = '';

/** @type {Set<string>} */
const keys = new Set();

/** @type {((type: string, data: any) => void) | null} */
let sendFn = null;

function recompute() {
	const right = keys.has('d') || keys.has('arrowright');
	const left = keys.has('a') || keys.has('arrowleft');
	const down = keys.has('s') || keys.has('arrowdown');
	const up = keys.has('w') || keys.has('arrowup');
	dx = (right ? 1 : 0) - (left ? 1 : 0);
	dy = (down ? 1 : 0) - (up ? 1 : 0);
	shoot = keys.has(' ');
}

function flush() {
	const cur = `${dx},${dy},${shoot}`;
	if (cur !== prev) {
		prev = cur;
		sendFn?.('input', { dx, dy, shoot });
	}
}

/** @param {KeyboardEvent} e @param {boolean} pressed */
function onKey(e, pressed) {
	const k = e.key.toLowerCase();
	if (['arrowup', 'arrowdown', 'arrowleft', 'arrowright', ' '].includes(k)) {
		e.preventDefault();
	}
	if (pressed) keys.add(k);
	else keys.delete(k);
	recompute();
	flush();
}

/** @param {KeyboardEvent} e */
const kd = (e) => onKey(e, true);
/** @param {KeyboardEvent} e */
const ku = (e) => onKey(e, false);
const blur = () => {
	keys.clear();
	recompute();
	flush();
};

/** @param {(type: string, data: any) => void} send */
export function startInput(send) {
	sendFn = send;
	keys.clear();
	dx = dy = 0;
	shoot = false;
	prev = '';
	window.addEventListener('keydown', kd);
	window.addEventListener('keyup', ku);
	window.addEventListener('blur', blur);
}

export function stopInput() {
	window.removeEventListener('keydown', kd);
	window.removeEventListener('keyup', ku);
	window.removeEventListener('blur', blur);
	sendFn?.('input', { dx: 0, dy: 0, shoot: false });
	sendFn = null;
}
