<script>
	import { onMount } from 'svelte';
	import { game, teamColor } from '$lib/stores/game.svelte.js';
	import { send } from '$lib/net/socket.svelte.js';

	const sb = $derived(game.scoreboard);
	let showButton = $state(false);

	/** @param {string} t */
	function teamName(t) {
		if (t === 'RED') return 'Vermelho';
		if (t === 'GREEN') return 'Verde';
		return t;
	}

	const winnerText = $derived(
		!sb ? '' : sb.winner === 'DRAW' ? 'Empate!' : `${teamName(sb.winner)} Venceu!`
	);
	const winnerColor = $derived(sb && sb.winner !== 'DRAW' ? teamColor(sb.winner) : '#fff');

	onMount(() => {
		const id = setTimeout(() => (showButton = true), 5000);
		return () => clearTimeout(id);
	});

	function back() {
		send('return_to_lobby');
	}
</script>

<div class="screen">
	{#if sb}
		<p class="score">{sb.scoreText}</p>
		<h1 style:color={winnerColor}>{winnerText}</h1>
		{#if showButton}
			<button onclick={back}>Voltar à Sala</button>
		{:else}
			<p class="wait">Retornando em breve…</p>
		{/if}
	{/if}
</div>

<style>
	.screen {
		min-height: 100vh;
		display: grid;
		place-content: center;
		justify-items: center;
		gap: 1rem;
		background: #404040;
		color: #fff;
		font-family: system-ui, sans-serif;
		text-align: center;
	}
	.score {
		font-size: 2rem;
		font-weight: 700;
		margin: 0;
		font-variant-numeric: tabular-nums;
	}
	h1 {
		font-size: 2.6rem;
		margin: 0;
	}
	button {
		padding: 0.7rem 2rem;
		font-size: 1rem;
		border: none;
		border-radius: 6px;
		background: #7ed321;
		color: #1a1a1a;
		font-weight: 700;
		cursor: pointer;
	}
	.wait {
		color: #aaa;
		min-height: 2.6rem;
	}
</style>
