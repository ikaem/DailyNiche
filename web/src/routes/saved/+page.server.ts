import { getFavoritedPosts, getReadLaterPosts } from '$lib/server/api';
import {
	favoritePostAction,
	markReadLaterAction,
	unfavoritePostAction,
	unmarkReadLaterAction
} from '$lib/server/savedPostActions';
import type { Actions, PageServerLoad } from './$types';

// Fetches both lists unconditionally, regardless of which tab is showing -
// the tab bar needs accurate counts for both tabs displayed at once (see
// saved-v1.html's design), and tab-switching itself is plain client-side
// state in +page.svelte, not a server round-trip, so there's no per-tab
// variant of this load to begin with.
export const load: PageServerLoad = async () => {
	try {
		const [favoritedPosts, readLaterPosts] = await Promise.all([
			getFavoritedPosts(),
			getReadLaterPosts()
		]);
		return { favoritedPosts, readLaterPosts, error: null };
	} catch (err) {
		return {
			favoritedPosts: [],
			readLaterPosts: [],
			error: err instanceof Error ? err.message : 'Failed to load saved posts'
		};
	}
};

// Same four actions as the home page (see $lib/server/savedPostActions) -
// a post can be un-favorited/un-read-later'd directly from this page too,
// same as from the home feed.
export const actions: Actions = {
	favoritePost: favoritePostAction,
	unfavoritePost: unfavoritePostAction,
	markReadLater: markReadLaterAction,
	unmarkReadLater: unmarkReadLaterAction
};
