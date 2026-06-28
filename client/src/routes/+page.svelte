<script>
	import { game } from '$lib/stores/game.svelte.js';
	import Login from '$lib/components/Login.svelte';
	import Lobby from '$lib/components/Lobby.svelte';
	import Game from '$lib/components/Game.svelte';

	const screen = $derived(
		game.connection !== 'connected'
			? 'login'
			: game.phase === 'PLAYING'
				? 'game'
				: game.phase === 'SPECTATOR'
					? 'spectator'
					: game.phase === 'SCOREBOARD'
						? 'scoreboard'
						: 'lobby'
	);
</script>

{#if screen === 'login'}
	<Login />
{:else if screen === 'lobby'}
	<Lobby />
{:else if screen === 'game'}
	<Game />
{:else}
	<div class="loading">Loading…</div>
{/if}

<style>
	.loading {
		min-height: 100vh;
		display: grid;
		place-items: center;
		background: #404040;
		color: #fff;
		font-family: system-ui, sans-serif;
	}
</style>
