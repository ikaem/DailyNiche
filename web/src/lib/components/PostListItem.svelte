<!--
	Below-the-fold list row: used for every post after the 2 PostHero +
	4 PostMedium posts. Single-column text row (with a small thumbnail),
	the lowest tier of visual weight.
-->
<script lang="ts">
	import type { PostModel } from '$lib/types';

	let { post }: { post: PostModel } = $props();
</script>

<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- post.url is an external article link, not an internal SvelteKit route -->
<a class="list-post" href={post.url} target="_blank" rel="noopener noreferrer">
	<img src={post.imageUrl} alt="" />
	<div class="list-post-text">
		<span class="kicker">{post.feedName}</span><span class="date-dot"
			>{post.publishedAtDisplay}</span
		>
		<h4>{post.title}</h4>
		<p class="desc">{post.description}</p>
	</div>
</a>

<style>
	.list-post {
		grid-column: 1 / -1;
		display: flex;
		align-items: center;
		gap: 1.1rem;
		padding: 0.9rem 0.5rem;
		border-radius: 10px;
		color: inherit;
		text-decoration: none;
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
