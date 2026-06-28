<script>
	import { game } from '$lib/stores/game.svelte.js';
	import { connect } from '$lib/net/socket.svelte.js';

	let username = $state('');
	const valid = $derived(/^[a-z]{3,16}$/.test(username));
	const connecting = $derived(game.connection === 'connecting');

	/** @param {SubmitEvent} e */
	function submit(e) {
		e.preventDefault();
		if (valid) connect(username);
	}
</script>

<div class="screen">
	<div class="card">
		<h1>Minimalist Paint War</h1>
		<form onsubmit={submit}>
			<input
				type="text"
				placeholder="username"
				bind:value={username}
				autocomplete="off"
				autocapitalize="none"
				spellcheck="false"
				maxlength="16"
			/>
			<p class="hint">3–16 lowercase letters only (a–z).</p>
			<button type="submit" disabled={!valid || connecting}>
				{connecting ? 'Connecting…' : 'Enter'}
			</button>
		</form>
		{#if game.error}
			<p class="error">{game.error.message}</p>
		{/if}
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
	.card {
		background: #2e2e2e;
		padding: 2rem;
		border-radius: 10px;
		width: min(360px, 90vw);
		text-align: center;
	}
	h1 {
		margin: 0 0 1.5rem;
		font-size: 1.4rem;
	}
	input {
		width: 100%;
		padding: 0.6rem;
		font-size: 1rem;
		border: none;
		border-radius: 6px;
		box-sizing: border-box;
	}
	.hint {
		font-size: 0.75rem;
		color: #aaa;
		margin: 0.4rem 0 1rem;
	}
	button {
		width: 100%;
		padding: 0.6rem;
		font-size: 1rem;
		border: none;
		border-radius: 6px;
		background: #7ed321;
		color: #1a1a1a;
		font-weight: 600;
		cursor: pointer;
	}
	button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.error {
		color: #e74c3c;
		margin: 1rem 0 0;
		font-size: 0.85rem;
	}
</style>
