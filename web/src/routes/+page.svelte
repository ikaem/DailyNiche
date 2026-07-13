<script lang="ts">
	import type { Post } from '$lib/types';
	import { toPostModel } from '$lib/postModel';
	import DateNav from '$lib/components/DateNav.svelte';
	import AboveTheFold from '$lib/components/AboveTheFold.svelte';
	import BelowTheFold from '$lib/components/BelowTheFold.svelte';

	// Hardcoded dummy data for now, matching docs/design/issue/issue-v6.html -
	// Task 6.1 swaps this for a real API call, through the same toPostModel()
	// mapping used here.
	const dummyPosts: Post[] = [
		{
			id: 1,
			title: 'Go 2.0 Announced',
			description:
				'The Go team announces the next major version, with a focus on generics and tooling improvements.',
			imageUrl:
				'https://images.unsplash.com/photo-1506744038136-46273834b3fb?w=1000&h=750&fit=crop',
			url: 'https://example.com/go-2-0-announced',
			feedName: 'Tech Blog',
			publishedAt: '2026-07-10T09:00:00Z'
		},
		{
			id: 2,
			title: 'Why SQLite Is Enough',
			description:
				"A case for boring, reliable databases - and why most side projects don't need more.",
			imageUrl:
				'https://images.unsplash.com/photo-1519389950473-47ba0277781c?w=1000&h=750&fit=crop',
			url: 'https://example.com/why-sqlite-is-enough',
			feedName: 'Tech Blog',
			publishedAt: '2026-07-10T07:30:00Z'
		},
		{
			id: 3,
			title: 'Perfect Sourdough Starter',
			description: 'A no-fuss guide to your first starter.',
			imageUrl: 'https://images.unsplash.com/photo-1517245386807-bb43f82c33c4?w=400&h=300&fit=crop',
			url: 'https://example.com/perfect-sourdough-starter',
			feedName: 'Cooking Blog',
			publishedAt: '2026-07-10T11:15:00Z'
		},
		{
			id: 4,
			title: 'A Weekend in Ljubljana',
			description: 'Small city, big charm.',
			imageUrl: 'https://images.unsplash.com/photo-1495020689067-958852a7765e?w=400&h=300&fit=crop',
			url: 'https://example.com/a-weekend-in-ljubljana',
			feedName: 'Travel Blog',
			publishedAt: '2026-07-10T10:00:00Z'
		},
		{
			id: 5,
			title: 'Static Sites Are Back',
			description: 'The pendulum swings again.',
			imageUrl: 'https://images.unsplash.com/photo-1504711434969-e33886168f5c?w=400&h=300&fit=crop',
			url: 'https://example.com/static-sites-are-back',
			feedName: 'Tech Blog',
			publishedAt: '2026-07-10T08:45:00Z'
		},
		{
			id: 6,
			title: 'Weeknight Pasta Ideas',
			description: 'Five pastas in under 20 minutes.',
			imageUrl: 'https://images.unsplash.com/photo-1498050108023-c5249f4df085?w=400&h=300&fit=crop',
			url: 'https://example.com/weeknight-pasta-ideas',
			feedName: 'Cooking Blog',
			publishedAt: '2026-07-10T06:20:00Z'
		},
		{
			id: 7,
			title: 'Packing Light for Winter',
			description:
				'One bag, three weeks, no regrets - a practical guide to minimalist winter travel.',
			imageUrl: 'https://images.unsplash.com/photo-1506744038136-46273834b3fb?w=200&h=150&fit=crop',
			url: 'https://example.com/packing-light-for-winter',
			feedName: 'Travel Blog',
			publishedAt: '2026-07-09T14:00:00Z'
		},
		{
			id: 8,
			title: 'Another Tech Blog Post Title Goes Here',
			description: 'A short summary of what this post is actually about, for context.',
			imageUrl: 'https://images.unsplash.com/photo-1519389950473-47ba0277781c?w=200&h=150&fit=crop',
			url: 'https://example.com/another-tech-blog-post',
			feedName: 'Tech Blog',
			publishedAt: '2026-07-09T09:30:00Z'
		},
		{
			id: 9,
			title: 'Yet Another Post From The Cooking Blog',
			description: 'A quick one-liner describing the recipe or article contents.',
			imageUrl: 'https://images.unsplash.com/photo-1517245386807-bb43f82c33c4?w=200&h=150&fit=crop',
			url: 'https://example.com/yet-another-cooking-post',
			feedName: 'Cooking Blog',
			publishedAt: '2026-07-08T16:45:00Z'
		},
		{
			id: 10,
			title: 'A Fourth Post To Show The List Continues',
			description: 'Placeholder description text, roughly the length a real summary would be.',
			imageUrl: 'https://images.unsplash.com/photo-1495020689067-958852a7765e?w=200&h=150&fit=crop',
			url: 'https://example.com/a-fourth-list-post',
			feedName: 'Travel Blog',
			publishedAt: '2026-07-08T12:10:00Z'
		}
	];

	const posts = dummyPosts.map(toPostModel);
</script>

<DateNav />

<main>
	<div class="grid-12">
		<AboveTheFold {posts} />
		<BelowTheFold {posts} />
	</div>
</main>

<style>
	main {
		max-width: 1200px;
		margin: 0 auto;
		padding: 1.5rem 2rem 3rem;
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
