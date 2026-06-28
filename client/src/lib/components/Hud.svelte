<script>
	import { game, teamColor } from '$lib/stores/game.svelte.js';

	let { spectator = false } = $props();

	const snap = $derived(game.snapshot);
	const timeText = $derived(formatTime(snap?.timeLeftMs ?? 0));
	const me = $derived(snap?.players?.find((/** @type {any} */ p) => p.id === game.me.id));
	const myFlags = $derived(snap?.flagCount?.[game.me.team] ?? 0);

	/** @param {number} ms */
	function formatTime(ms) {
		const s = Math.max(0, Math.ceil(ms / 1000));
		const m = Math.floor(s / 60);
		return `${String(m).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`;
	}
</script>

<div class="timer">{timeText}</div>

<div class="flags">
	<span style:color={teamColor('RED')}>{snap?.flagCount?.RED ?? 0}</span>
	<span class="sep">:</span>
	<span style:color={teamColor('GREEN')}>{snap?.flagCount?.GREEN ?? 0}</span>
</div>

{#if !spectator}
	<div class="lives">
		<span class="label">Vidas</span>
		{#each [0, 1, 2] as i (i)}
			<span class="hp" class:on={me && me.hp > i}></span>
		{/each}
		<span class="flagcount" style:color={teamColor(game.me.team)}>{myFlags} ⚑</span>
	</div>
{/if}

<style>
	.timer {
		position: absolute;
		top: 12px;
		left: 50%;
		transform: translateX(-50%);
		background: #000;
		color: #fff;
		font-weight: 700;
		font-size: 1.4rem;
		padding: 0.25rem 0.9rem;
		border-radius: 8px;
		font-variant-numeric: tabular-nums;
	}
	.flags {
		position: absolute;
		top: 16px;
		right: 16px;
		font-weight: 700;
		font-size: 1.3rem;
		background: rgba(0, 0, 0, 0.45);
		padding: 0.2rem 0.7rem;
		border-radius: 8px;
	}
	.sep {
		color: #fff;
	}
	.lives {
		position: absolute;
		bottom: 14px;
		right: 16px;
		display: flex;
		align-items: center;
		gap: 6px;
		color: #fff;
		font-weight: 700;
	}
	.hp {
		width: 14px;
		height: 14px;
		border-radius: 50%;
		background: #777;
		display: inline-block;
	}
	.hp.on {
		background: #e74c3c;
	}
	.flagcount {
		margin-left: 8px;
		font-size: 1.2rem;
	}
</style>
