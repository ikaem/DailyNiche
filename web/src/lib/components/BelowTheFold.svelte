<!--
	Below-the-fold section: every post after the day's top 6 (2 hero + 4
	medium from AboveTheFold), rendered as PostListItem rows under an
	"Also today" divider. No wrapping element here - renders directly
	into the caller's `.grid-12` container (see +page.svelte), matching
	AboveTheFold's approach.
-->
<script lang="ts">
	import type { PostModel } from '$lib/types';
	import PostListItem from './PostListItem.svelte';

	let { posts }: { posts: PostModel[] } = $props();

	let belowFoldPosts = $derived(posts.slice(6));
</script>

{#if belowFoldPosts.length > 0}
	<div class="briefs-label">Also today</div>
	{#each belowFoldPosts as post (post.id)}
		<PostListItem {post} />
	{/each}
{/if}

<style>
	.briefs-label {
		grid-column: 1 / -1;
		font-family: 'Playfair Display', Georgia, serif;
		font-style: italic;
		font-size: 1.1rem;
		color: var(--ink-soft);
		border-bottom: 1px solid var(--line);
		padding-bottom: 0.6rem;
		margin-top: 0.5rem;
	}
</style>
