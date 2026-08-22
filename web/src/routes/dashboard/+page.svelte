<script lang="ts">
	import { enhance } from '$app/forms';
	import FeedRow from '$lib/components/FeedRow.svelte';
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let activeFeeds = $derived(data.feeds.filter((feed) => feed.disabledAt === null));
	let disabledFeeds = $derived(data.feeds.filter((feed) => feed.disabledAt !== null));

	// Tracks the in-flight state of the fetch-now form specifically (not a
	// generic "any form is submitting" flag) - a custom use:enhance callback
	// is needed here (the other forms on this page use enhance's bare
	// default) because we want the button's own label/disabled state to
	// reflect this one action's progress, which the default behavior alone
	// doesn't expose.
	let fetching = $state(false);
</script>

<main>
	<h1>Dashboard</h1>

	<section class="fetch-now">
		<form
			method="POST"
			action="?/fetchNow"
			use:enhance={() => {
				fetching = true;
				return async ({ update }) => {
					fetching = false;
					await update();
				};
			}}
		>
			<button type="submit" class="secondary" disabled={fetching}>
				{fetching ? 'Fetching…' : 'Fetch now'}
			</button>
		</form>
	</section>

	{#if data.error}
		<p class="status status-error">{data.error}</p>
	{:else}
		{#if form?.message}
			<p class="status status-error">{form.message}</p>
		{/if}
		{#if form?.fetchSummary}
			<p class="status status-success">{form.fetchSummary}</p>
		{/if}

		<section>
			<h2>Add a new feed</h2>
			<form method="POST" action="?/addFeed" use:enhance class="add-feed-form">
				<div class="field">
					<label for="name">Name</label>
					<input type="text" id="name" name="name" placeholder="e.g. Tech Blog" required />
				</div>
				<div class="field">
					<label for="url">URL</label>
					<input
						type="url"
						id="url"
						name="url"
						placeholder="https://example.com/feed.xml"
						required
					/>
				</div>
				<div class="form-actions">
					<button type="submit" class="primary">Add Feed</button>
				</div>
			</form>
		</section>

		<section>
			<h2>Active Feeds ({activeFeeds.length})</h2>
			{#each activeFeeds as feed (feed.id)}
				<FeedRow {feed} />
			{/each}
		</section>

		{#if disabledFeeds.length > 0}
			<section>
				<h2>Disabled Feeds ({disabledFeeds.length})</h2>
				{#each disabledFeeds as feed (feed.id)}
					<FeedRow {feed} />
				{/each}
			</section>
		{/if}
	{/if}
</main>

<style>
	main {
		max-width: 760px;
		margin: 0 auto;
		padding: 2.5rem 2rem 4rem;
	}

	main > h1 {
		font-size: 1.9rem;
		margin: 0 0 1.75rem;
	}

	.fetch-now {
		margin-bottom: 1.5rem;
	}
	button.secondary {
		background: none;
		border: 1px solid var(--line);
		border-radius: 6px;
		padding: 0.5rem 1.1rem;
		font-family: inherit;
		font-size: 0.85rem;
		font-weight: 500;
		color: var(--ink);
		cursor: pointer;
	}
	button.secondary:hover:not(:disabled) {
		background: var(--card);
	}
	button.secondary:disabled {
		cursor: default;
		opacity: 0.6;
	}

	.status {
		text-align: center;
		padding: 1rem;
		margin-bottom: 1.5rem;
	}
	.status-error {
		color: var(--accent);
	}
	.status-success {
		color: var(--ink-soft);
	}

	section {
		margin-bottom: 2.5rem;
	}
	section > h2 {
		font-size: 1rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--ink-soft);
		margin: 0 0 0.9rem;
	}

	.add-feed-form {
		border: 1px solid var(--line);
		border-radius: 8px;
		padding: 1.25rem 1.5rem;
		background: var(--card);
	}
	.field {
		margin-bottom: 1rem;
	}
	.field label {
		display: block;
		font-size: 0.85rem;
		font-weight: 500;
		margin-bottom: 0.35rem;
	}
	.field input {
		width: 100%;
		padding: 0.55rem 0.7rem;
		border: 1px solid var(--line);
		border-radius: 6px;
		font-family: inherit;
		font-size: 0.95rem;
		background: var(--paper);
		color: var(--ink);
	}
	.field input:focus {
		outline: none;
		border-color: var(--accent);
	}
	.form-actions {
		display: flex;
		justify-content: flex-end;
	}
	button.primary {
		background: var(--accent);
		color: #fff;
		border: none;
		border-radius: 6px;
		padding: 0.6rem 1.4rem;
		font-size: 0.9rem;
		font-weight: 600;
		cursor: pointer;
	}
	button.primary:hover {
		opacity: 0.92;
	}

	@media (max-width: 640px) {
		main {
			padding: 1.5rem 1.25rem 3rem;
		}
		main > h1 {
			font-size: 1.5rem;
		}
	}
</style>
