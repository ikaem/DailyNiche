<!--
	Below-the-fold list row: used for every post after the 2 PostHero +
	4 PostMedium posts. Single-column text row (with a small thumbnail),
	the lowest tier of visual weight.
-->
<script lang="ts">
	import SavedBadges from './SavedBadges.svelte';
	import type { PostModel } from '$lib/types';

	let { post }: { post: PostModel } = $props();
</script>

<div class="list-post">
	<img src={post.imageUrl} alt="" />
	<div class="list-post-text">
		<span class="kicker">{post.feedName}</span><span class="date-dot"
			>{post.publishedAtDisplay}</span
		>
		<h4>
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- post.url is an external article link, not an internal SvelteKit route -->
			<a href={post.url} target="_blank" rel="noopener noreferrer">{post.title}</a>
		</h4>
		<p class="desc">{post.description}</p>
	</div>
	<SavedBadges {post} overlay={false} />
</div>

<style>
	.list-post {
		grid-column: 1 / -1;
		display: flex;
		align-items: center;
		gap: 1.1rem;
		padding: 0.9rem 0.5rem;
		border-radius: 10px;
	}
	.list-post:hover {
		background: var(--card);
	}
	.list-post img {
		width: 84px;
		height: 60px;
		object-fit: cover;
		border-radius: 8px;
		flex-shrink: 0;
	}
	.list-post-text {
		/* Without this, a flex item shrinks to fit its own content by
		   default, so SavedBadges (the next flex item) would hug wherever
		   the text happens to end instead of sitting at the row's right
		   edge. flex: 1 makes this fill the remaining space instead;
		   min-width: 0 overrides flex's own default min-width: auto, which
		   would otherwise stop the description from wrapping/truncating
		   within that now-constrained width. */
		flex: 1;
		min-width: 0;
	}
	.list-post .kicker {
		display: inline-block;
		font-size: 0.7rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--accent);
		background: var(--accent-soft);
		padding: 0.2rem 0.6rem;
		border-radius: 999px;
		margin-bottom: 0.3rem;
	}
	.list-post .date-dot {
		color: var(--ink-soft);
		font-size: 0.8rem;
		font-weight: 500;
		margin-left: 0.5rem;
	}
	.list-post-text h4 {
		font-family: 'Playfair Display', Georgia, serif;
		font-size: 1.05rem;
		margin: 0 0 0.2rem;
		font-weight: 700;
	}
	.list-post-text h4 a {
		color: inherit;
		text-decoration: none;
	}
	.list-post-text h4 a:hover {
		text-decoration: underline;
	}
	.list-post-text p.desc {
		font-size: 0.85rem;
		color: var(--ink-soft);
		margin: 0 0 0.3rem;
	}

	@media (max-width: 640px) {
		.list-post img {
			width: 64px;
			height: 48px;
		}
	}
</style>
