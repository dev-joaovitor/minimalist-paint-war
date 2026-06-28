<script>
	import { onMount } from 'svelte';
	import { game } from '$lib/stores/game.svelte.js';
	import { send } from '$lib/net/socket.svelte.js';
	import { draw } from '$lib/game/renderer.js';
	import { startInput, stopInput } from '$lib/game/input.js';
	import Hud from './Hud.svelte';

	let { spectator = false } = $props();

	/** @type {HTMLCanvasElement} */
	let canvas;

	onMount(() => {
		const ctx = canvas.getContext('2d');
		if (!spectator) startInput(send);

		let raf = 0;
		const loop = () => {
			if (ctx && game.map) draw(ctx, game.map, game.snapshot, game.me.id);
			raf = requestAnimationFrame(loop);
		};
		raf = requestAnimationFrame(loop);

		return () => {
			cancelAnimationFrame(raf);
			if (!spectator) stopInput();
		};
	});
</script>

<div class="stage">
	<canvas bind:this={canvas} width={game.map?.width ?? 1000} height={game.map?.height ?? 760}
	></canvas>
	<Hud {spectator} />
	{#if spectator}
		<div class="spec-banner">Assistindo — aguarde a próxima partida</div>
	{/if}
</div>

<style>
	.stage {
		position: relative;
		min-height: 100vh;
		display: grid;
		place-items: center;
		background: #1a1a1a;
		overflow: hidden;
	}
	canvas {
		max-width: 100vw;
		max-height: 100vh;
		display: block;
	}
	.spec-banner {
		position: absolute;
		top: 12px;
		left: 16px;
		background: rgba(0, 0, 0, 0.6);
		color: #fff;
		padding: 0.4rem 0.8rem;
		border-radius: 8px;
		font-family: system-ui, sans-serif;
	}
</style>
