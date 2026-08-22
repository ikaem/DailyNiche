<!--
	Favorited/read-later badges shown on a post card - shared by
	PostHero/PostMedium/PostListItem so the icons and visibility logic
	aren't repeated in each. size="lg" is PostHero's larger treatment;
	the default (unset) matches PostMedium/PostListItem's smaller one.
	Renders nothing at all when neither state is set.

	overlay (default true) floats the badges via `position: absolute` in
	the parent's top-right corner - the parent needs `position: relative`
	on its own top-level element for this (true for PostHero/PostMedium,
	which both have a large image to float over). PostListItem has no such
	image, so it passes overlay={false} to get a plain inline flex item
	instead, placed wherever the parent's own layout puts it.
-->
<script lang="ts">
	import type { PostModel } from '$lib/types';

	let {
		post,
		size = 'md',
		overlay = true
	}: { post: PostModel; size?: 'lg' | 'md'; overlay?: boolean } = $props();
</script>

{#if post.isFavorited || post.isReadLater}
	<div class="badges" class:lg={size === 'lg'} class:overlay>
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

<style>
	.badges {
		display: flex;
		gap: 0.35rem;
		flex-shrink: 0;
	}
	.badges.overlay {
		position: absolute;
		top: 0.6rem;
		right: 0.6rem;
		z-index: 2;
	}
	.badges.overlay.lg {
		top: 0.75rem;
		right: 0.75rem;
		gap: 0.4rem;
	}
	.badge {
		width: 26px;
		height: 26px;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.92);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--accent);
		box-shadow: 0 1px 4px rgba(24, 20, 15, 0.15);
	}
	.badges.lg .badge {
		width: 30px;
		height: 30px;
	}
	.badge svg {
		width: 14px;
		height: 14px;
	}
	.badges.lg .badge svg {
		width: 16px;
		height: 16px;
	}
</style>
