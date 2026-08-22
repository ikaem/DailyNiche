<script lang="ts">
	import { toPostModel } from '$lib/postModel';
	import PostListItem from '$lib/components/PostListItem.svelte';
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	// Client-side only, not a ?tab= query param - both lists are already
	// loaded together (see +page.server.ts), so switching tabs never needs
	// a server round-trip, matching how saved-v1.html's mockup behaves.
	let activeTab = $state<'favorites' | 'read-later'>('favorites');

	let favoritedPosts = $derived(data.favoritedPosts.map(toPostModel));
	let readLaterPosts = $derived(data.readLaterPosts.map(toPostModel));
</script>

<main>
	<h1>Saved</h1>

	{#if data.error}
		<p class="status status-error">{data.error}</p>
	{:else}
		{#if form?.message}
			<p class="status status-error">{form.message}</p>
		{/if}

		<div class="tabs">
			<button
				type="button"
				class="tab-btn"
				class:active={activeTab === 'favorites'}
				onclick={() => (activeTab = 'favorites')}
			>
				<svg viewBox="0 0 24 24" fill="currentColor"
					><path
						d="M12 21s-7.5-4.6-10-9C.4 8.5 2 4 6.2 4 8.7 4 10.7 5.5 12 7.3 13.3 5.5 15.3 4 17.8 4 22 4 23.6 8.5 22 12c-2.5 4.4-10 9-10 9z"
					/></svg
				>
				Favorites <span class="count">({favoritedPosts.length})</span>
			</button>
			<button
				type="button"
				class="tab-btn"
				class:active={activeTab === 'read-later'}
				onclick={() => (activeTab = 'read-later')}
			>
				<svg viewBox="0 0 24 24" fill="currentColor"
					><path d="M6 2a1 1 0 0 0-1 1v19l7-4 7 4V3a1 1 0 0 0-1-1H6z" /></svg
				>
				Read Later <span class="count">({readLaterPosts.length})</span>
			</button>
		</div>

		{#if activeTab === 'favorites'}
			{#if favoritedPosts.length === 0}
				<p class="empty-state">No favorited posts yet.</p>
			{:else}
				{#each favoritedPosts as post (post.id)}
					<PostListItem {post} />
				{/each}
			{/if}
		{:else if readLaterPosts.length === 0}
			<p class="empty-state">No posts saved for later yet.</p>
		{:else}
			{#each readLaterPosts as post (post.id)}
				<PostListItem {post} />
			{/each}
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
		margin: 0 0 1.5rem;
	}

	.status {
		text-align: center;
		padding: 1rem;
		margin-bottom: 1.5rem;
	}
	.status-error {
		color: var(--accent);
	}

	/* Pill-style tabs, matching date-nav/kicker's rounded visual language -
	   see docs/design/saved/saved-v1.html. */
	.tabs {
		display: inline-flex;
		background: var(--card);
		border: 1px solid var(--line);
		border-radius: 999px;
		padding: 0.3rem;
		gap: 0.3rem;
		margin-bottom: 1.5rem;
		box-shadow: 0 1px 3px rgba(24, 20, 15, 0.05);
	}
	.tab-btn {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		border: none;
		background: none;
		border-radius: 999px;
		padding: 0.5rem 1.1rem;
		font-family: inherit;
		font-size: 0.88rem;
		font-weight: 600;
		color: var(--ink-soft);
		cursor: pointer;
	}
	.tab-btn svg {
		width: 15px;
		height: 15px;
		flex-shrink: 0;
	}
	.tab-btn.active {
		background: var(--accent);
		color: #fff;
	}
	.tab-btn:not(.active):hover {
		background: var(--accent-soft);
		color: var(--accent);
	}
	.tab-btn .count {
		opacity: 0.75;
		font-weight: 500;
	}

	.empty-state {
		text-align: center;
		padding: 3rem 1rem;
		color: var(--ink-soft);
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
