<script lang="ts">
	import { toPostModel } from '$lib/postModel';
	import DateNav from '$lib/components/DateNav.svelte';
	import AboveTheFold from '$lib/components/AboveTheFold.svelte';
	import BelowTheFold from '$lib/components/BelowTheFold.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let posts = $derived(data.posts.map(toPostModel));
</script>

<DateNav />

<main>
	{#if data.error}
		<p class="status status-error">{data.error}</p>
	{:else}
		<div class="grid-12">
			<AboveTheFold {posts} />
			<BelowTheFold {posts} />
		</div>
	{/if}
</main>

<style>
	main {
		max-width: 1200px;
		margin: 0 auto;
		padding: 1.5rem 2rem 3rem;
	}

	.status {
		text-align: center;
		padding: 4rem 1rem;
		color: var(--ink-soft);
	}
	.status-error {
		color: var(--accent);
	}

	.grid-12 {
		display: grid;
		grid-template-columns: repeat(12, 1fr);
		gap: 1.75rem;
	}

	@media (max-width: 640px) {
		main {
			padding: 1rem 1rem 2rem;
		}
		.grid-12 {
			gap: 1rem;
		}
	}
</style>
