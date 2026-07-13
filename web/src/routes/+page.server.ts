import { getPostsToday } from '$lib/server/api';
import type { PageServerLoad } from './$types';

// Runs only on the server - the browser never talks to the Go API directly,
// so it's never subject to CORS. Errors are returned as data (not thrown)
// so +page.svelte can show an inline message without losing the header/nav.
export const load: PageServerLoad = async () => {
	try {
		const posts = await getPostsToday();
		return { posts, error: null };
	} catch (err) {
		return { posts: [], error: err instanceof Error ? err.message : 'Failed to load posts' };
	}
};
