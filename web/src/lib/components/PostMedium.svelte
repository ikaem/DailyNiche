<!--
	Above-the-fold medium card: used for the next 4 posts of the day
	(after the 2 PostHero posts). Plain card - image on top, content
	below - a step down in visual weight from PostHero.
-->
<script lang="ts">
	import type { PostModel } from '$lib/types';

	let { post }: { post: PostModel } = $props();
</script>

<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- post.url is an external article link, not an internal SvelteKit route -->
<a class="medium-post" href={post.url} target="_blank" rel="noopener noreferrer">
	<img src={post.imageUrl} alt="" />
	<div class="content">
		<span class="kicker">{post.feedName}</span>
		<h3>{post.title}</h3>
		<p class="desc">{post.description}</p>
	</div>
</a>

<style>
	.medium-post {
		grid-column: span 3;
		background: var(--card);
		border-radius: 12px;
		overflow: hidden;
		box-shadow: 0 2px 10px rgba(24, 20, 15, 0.06);
		border: 1px solid var(--line);
		display: block;
		color: inherit;
		text-decoration: none;
	}
	.medium-post img {
		width: 100%;
		height: 150px;
		object-fit: cover;
		display: block;
	}
	.medium-post .content {
		padding: 0.9rem 1rem 1.1rem;
	}
	.medium-post .kicker {
		display: inline-block;
		font-size: 0.7rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--accent);
		background: var(--accent-soft);
		padding: 0.2rem 0.6rem;
		border-radius: 999px;
		margin-bottom: 0.6rem;
	}
	.medium-post h3 {
		font-family: 'Playfair Display', Georgia, serif;
		font-size: 1.05rem;
		line-height: 1.25;
		margin: 0 0 0.3rem;
	}
	.medium-post p.desc {
		font-family: 'Inter', Arial, sans-serif;
		font-size: 0.82rem;
		color: var(--ink-soft);
		margin: 0;
	}

	@media (max-width: 640px) {
		.medium-post {
			grid-column: span 6;
		}
		.medium-post img {
			height: 110px;
		}
	}
</style>
