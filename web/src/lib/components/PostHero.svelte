<script lang="ts">
	import type { PostModel } from '$lib/types';

	let { post }: { post: PostModel } = $props();
</script>

<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- post.url is an external article link, not an internal SvelteKit route -->
<a class="hero-post" href={post.url} target="_blank" rel="noopener noreferrer">
	<img src={post.imageUrl} alt="" />
	<div class="content">
		<span class="kicker">{post.feedName}</span>
		<h2>{post.title}</h2>
		<p class="desc">{post.description}</p>
		<span class="date-dot">{post.publishedAtDisplay}</span>
	</div>
</a>

<style>
	.hero-post {
		grid-column: span 6;
		position: relative;
		border-radius: 16px;
		overflow: hidden;
		box-shadow: 0 8px 24px rgba(24, 20, 15, 0.12);
		color: #fff;
		min-height: 360px;
		display: flex;
		align-items: flex-end;
		text-decoration: none;
	}
	.hero-post img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		z-index: 0;
	}
	.hero-post::after {
		content: '';
		position: absolute;
		inset: 0;
		background: linear-gradient(
			to top,
			rgba(10, 8, 6, 0.88) 0%,
			rgba(10, 8, 6, 0.35) 55%,
			rgba(10, 8, 6, 0) 100%
		);
		z-index: 1;
	}
	.hero-post .content {
		position: relative;
		z-index: 2;
		padding: 1.5rem;
	}
	.hero-post .kicker {
		display: inline-block;
		font-size: 0.7rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		padding: 0.2rem 0.6rem;
		border-radius: 999px;
		margin-bottom: 0.6rem;
		background: rgba(255, 255, 255, 0.15);
		color: #fff;
		backdrop-filter: blur(4px);
	}
	.hero-post h2 {
		font-family: 'Playfair Display', Georgia, serif;
		font-size: 1.7rem;
		line-height: 1.15;
		margin: 0 0 0.4rem;
	}
	.hero-post p.desc {
		font-size: 0.95rem;
		color: rgba(255, 255, 255, 0.88);
		margin: 0 0 0.5rem;
		max-width: 46ch;
	}
	.hero-post .date-dot {
		color: rgba(255, 255, 255, 0.7);
		font-size: 0.8rem;
		font-weight: 500;
	}

	@media (max-width: 640px) {
		.hero-post {
			grid-column: span 12;
			min-height: 260px;
		}
		.hero-post h2 {
			font-size: 1.3rem;
		}
	}
</style>
