<script>
	import { game } from '$lib/stores/game.svelte.js';
	const rows = $derived(game.lobby.leaderboard ?? []);
</script>

<div class="board">
	<h3>Ranking</h3>
	{#if rows.length === 0}
		<p class="empty">Nenhuma partida jogada ainda.</p>
	{:else}
		<table>
			<thead>
				<tr><th>#</th><th>Jogador</th><th>V</th><th>D</th></tr>
			</thead>
			<tbody>
				{#each rows as r, i (r.username)}
					<tr>
						<td>{i + 1}</td>
						<td>{r.username}</td>
						<td>{r.wins}</td>
						<td>{r.losses}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</div>

<style>
	.board {
		background: #2e2e2e;
		border-radius: 8px;
		padding: 1rem;
	}
	h3 {
		margin: 0 0 0.6rem;
		font-size: 1rem;
	}
	.empty {
		color: #aaa;
		font-size: 0.85rem;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85rem;
	}
	th,
	td {
		text-align: left;
		padding: 0.25rem 0.4rem;
	}
	th:nth-child(3),
	th:nth-child(4),
	td:nth-child(3),
	td:nth-child(4) {
		text-align: right;
	}
	tbody tr:nth-child(odd) {
		background: #383838;
	}
</style>
