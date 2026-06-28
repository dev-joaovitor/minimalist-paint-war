<script>
	import { game, teamColor } from '$lib/stores/game.svelte.js';
	import { send, disconnect } from '$lib/net/socket.svelte.js';
	import Leaderboard from './Leaderboard.svelte';

	const players = $derived(game.lobby.players ?? []);
	const canStart = $derived(game.isLeader && players.length >= 2);

	function start() {
		send('start_game');
	}
</script>

<div class="screen">
	<div class="panel">
		<header>
			<h1>Sala</h1>
			<button class="leave" onclick={disconnect}>Sair</button>
		</header>

		<div class="grid">
			<div class="players">
				<h3>Jogadores ({players.length})</h3>
				<ul>
					{#each players as p (p.id)}
						<li>
							<span class="dot" style:background={teamColor(p.team)}></span>
							<span class="name">{p.username}</span>
							{#if p.id === game.lobby.leaderId}<span class="badge">líder</span>{/if}
							{#if p.id === game.me.id}<span class="you">você</span>{/if}
						</li>
					{/each}
				</ul>
			</div>
			<Leaderboard />
		</div>

		<footer>
			{#if game.isLeader}
				<button class="start" onclick={start} disabled={!canStart}>Iniciar</button>
				{#if !canStart}<p class="hint">São necessários ao menos 2 jogadores.</p>{/if}
			{:else}
				<p class="hint">Aguardando o líder iniciar…</p>
			{/if}
			{#if game.error}<p class="error">{game.error.message}</p>{/if}
		</footer>
	</div>
</div>

<style>
	.screen {
		min-height: 100vh;
		display: grid;
		place-items: center;
		background: #404040;
		color: #fff;
		font-family: system-ui, sans-serif;
	}
	.panel {
		width: min(720px, 94vw);
		background: #2e2e2e;
		border-radius: 10px;
		padding: 1.5rem;
	}
	header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}
	h1 {
		margin: 0;
		font-size: 1.4rem;
	}
	h3 {
		margin: 0 0 0.6rem;
		font-size: 1rem;
	}
	.grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}
	@media (max-width: 560px) {
		.grid {
			grid-template-columns: 1fr;
		}
	}
	.players {
		background: #383838;
		border-radius: 8px;
		padding: 1rem;
	}
	ul {
		list-style: none;
		margin: 0;
		padding: 0;
	}
	li {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.3rem 0;
	}
	.dot {
		width: 12px;
		height: 12px;
		border-radius: 3px;
		display: inline-block;
	}
	.name {
		flex: 1;
	}
	.badge {
		font-size: 0.7rem;
		background: #f1c40f;
		color: #1a1a1a;
		padding: 0.1rem 0.4rem;
		border-radius: 4px;
	}
	.you {
		font-size: 0.7rem;
		color: #aaa;
	}
	footer {
		margin-top: 1.2rem;
		text-align: center;
	}
	.start {
		padding: 0.7rem 2rem;
		font-size: 1rem;
		border: none;
		border-radius: 6px;
		background: #7ed321;
		color: #1a1a1a;
		font-weight: 700;
		cursor: pointer;
	}
	.start:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.leave {
		background: none;
		border: 1px solid #666;
		color: #ccc;
		padding: 0.3rem 0.8rem;
		border-radius: 6px;
		cursor: pointer;
	}
	.hint {
		color: #aaa;
		font-size: 0.85rem;
		margin: 0.6rem 0 0;
	}
	.error {
		color: #e74c3c;
		font-size: 0.85rem;
		margin: 0.6rem 0 0;
	}
</style>
