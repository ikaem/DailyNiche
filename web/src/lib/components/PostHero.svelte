<!--
	Above-the-fold hero card: used for the top 2 posts of the day.
	Image-overlay treatment (title/desc sit on top of the photo) marks
	these as the most prominent posts. See PostMedium for the next tier
	down and PostListItem for everything below the fold.
-->
<script lang="ts">
	import type { PostModel } from '$lib/types';

	let { post }: { post: PostModel } = $props();
</script>

<article class="hero-post">
	<img src={post.imageUrl} alt="" />
	{#if post.isFavorited || post.isReadLater}
		<div class="badges">
			{#if post.isFavorited}
				<span class="badge" title="Favorited">
					<svg viewBox="0 0 24 24" fill="currentColor"
						><path
							d="M12 21s-7.5-4.6-10-9C.4 8.5 2 4 6.2 4 8.7 4 10.7 5.5 12 7.3 13.3 5.5 15.3 4 17.8 4 22 4 23.6 8.5 22 12c-2.5 4.4-10 9-10 9z"
						/></svg
					>
				</span>
			{/if}
			{#if post.isReadLater}
				<span class="badge" title="Read later">
					<svg viewBox="0 0 24 24" fill="currentColor"
						><path d="M6 2a1 1 0 0 0-1 1v19l7-4 7 4V3a1 1 0 0 0-1-1H6z" /></svg
					>
				</span>
			{/if}
		</div>
	{/if}
	<div class="content">
		<span class="kicker">{post.feedName}</span>
		<h2>
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- post.url is an external article link, not an internal SvelteKit route -->
			<a href={post.url} target="_blank" rel="noopener noreferrer">{post.title}</a>
		</h2>
		<p class="desc">{post.description}</p>
		<span class="date-dot">{post.publishedAtDisplay}</span>
	</div>
</article>

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
	.hero-post h2 a {
		color: inherit;
		text-decoration: none;
	}
	.hero-post h2 a:hover {
		text-decoration: underline;
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

	/* Favorited/read-later badges - matching docs/design/saved/saved-v1.html's
	   icon-badge treatment: a white circle so the icon stays legible against
	   any photo, positioned above the gradient scrim (z-index 2, same as
	   .content). */
	.hero-post .badges {
		position: absolute;
		top: 0.75rem;
		right: 0.75rem;
		z-index: 2;
		display: flex;
		gap: 0.4rem;
	}
	.hero-post .badge {
		width: 30px;
		height: 30px;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.92);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--accent);
		box-shadow: 0 1px 4px rgba(24, 20, 15, 0.15);
	}
	.hero-post .badge svg {
		width: 16px;
		height: 16px;
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
