import type { Post, PostModel } from './types';

// toPostModel maps a raw Post (as delivered by the API client) into a
// PostModel ready for rendering - the only place date formatting happens,
// rather than every component doing its own conversion.
export function toPostModel(post: Post): PostModel {
	return {
		id: post.id,
		title: post.title,
		description: post.description,
		imageUrl: post.imageUrl,
		url: post.url,
		feedName: post.feedName,
		publishedAtDisplay: formatDate(post.publishedAt)
	};
}

// TODO: locale is hardcoded to Croatian for now. Should eventually be
// derived from the user's actual browser/location settings instead of a
// fixed value - see CLAUDE.md.
function formatDate(iso: string): string {
	return new Intl.DateTimeFormat('hr-HR', {
		month: 'short',
		day: 'numeric',
		year: 'numeric'
	}).format(new Date(iso));
}
