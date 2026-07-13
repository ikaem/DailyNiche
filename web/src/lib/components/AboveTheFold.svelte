<!--
	Above-the-fold section: the day's 2 most prominent posts as PostHero,
	followed by the next 4 as PostMedium. No wrapping element here - this
	renders directly into the caller's `.grid-12` container (see
	+page.svelte) so PostHero/PostMedium's grid-column spans apply
	correctly as direct grid children.
-->
<script lang="ts">
	import type { PostModel } from '$lib/types';
	import PostHero from './PostHero.svelte';
	import PostMedium from './PostMedium.svelte';

	let { posts }: { posts: PostModel[] } = $props();

	let heroPosts = $derived(posts.slice(0, 2));
	let mediumPosts = $derived(posts.slice(2, 6));
</script>

{#each heroPosts as post (post.id)}
	<PostHero {post} />
{/each}
{#each mediumPosts as post (post.id)}
	<PostMedium {post} />
{/each}
