<script lang="ts">
	import { enhance } from '$app/forms';
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let activeFeeds = $derived(data.feeds.filter((feed) => feed.disabledAt === null));
	let disabledFeeds = $derived(data.feeds.filter((feed) => feed.disabledAt !== null));
</script>

<main>
	<h1>Dashboard</h1>

	{#if data.error}
		<p class="status status-error">{data.error}</p>
	{:else}
		{#if form?.message}
			<p class="status status-error">{form.message}</p>
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
				<div class="feed-row">
					<div class="feed-info">
						<span class="feed-name">{feed.name}</span>
						<span class="feed-url">{feed.url}</span>
					</div>
					<form method="POST" action="?/deleteFeed" use:enhance>
						<input type="hidden" name="id" value={feed.id} />
						<button type="submit" class="delete">Delete</button>
					</form>
				</div>
			{/each}
		</section>

		{#if disabledFeeds.length > 0}
			<section>
				<h2>Disabled Feeds ({disabledFeeds.length})</h2>
				{#each disabledFeeds as feed (feed.id)}
					<div class="feed-row">
						<div class="feed-info">
							<span class="feed-name">{feed.name}</span>
							<span class="feed-url">{feed.url}</span>
						</div>
						<!-- Not wired to anything yet - repos.EnableFeed doesn't exist,
						     see the Task 7.4 TODO in CLAUDE.md. Looks and behaves like a
						     real button, just does nothing when clicked. -->
						<button type="button" class="enable">Enable</button>
					</div>
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

	.status {
		text-align: center;
		padding: 1rem;
		margin-bottom: 1.5rem;
	}
	.status-error {
		color: var(--accent);
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

	.feed-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
		padding: 0.85rem 0;
		border-bottom: 1px solid var(--line);
	}
	.feed-row:last-child {
		border-bottom: none;
	}
	.feed-row .feed-info {
		min-width: 0;
	}
	.feed-row .feed-name {
		font-weight: 600;
		font-size: 0.95rem;
	}
	.feed-row .feed-url {
		font-size: 0.8rem;
		color: var(--ink-soft);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		display: block;
	}
	.feed-row button {
		flex-shrink: 0;
		background: none;
		border: 1px solid var(--line);
		border-radius: 6px;
		padding: 0.4rem 0.9rem;
		font-family: inherit;
		font-size: 0.82rem;
		font-weight: 500;
		cursor: pointer;
		color: var(--ink);
	}
	.feed-row button.delete {
		color: var(--accent);
		border-color: var(--accent-soft);
	}
	.feed-row button.delete:hover {
		background: var(--accent-soft);
	}
	.feed-row button.enable:hover {
		background: var(--paper);
	}

	@media (max-width: 640px) {
		main {
			padding: 1.5rem 1.25rem 3rem;
		}
		main > h1 {
			font-size: 1.5rem;
		}

		.feed-row {
			flex-wrap: wrap;
		}
		.feed-row .feed-info {
			flex: 1 1 100%;
		}
	}
</style>
