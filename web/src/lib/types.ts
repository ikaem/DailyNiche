// Post matches exactly what the API client delivers - raw, unformatted.
// publishedAt stays a plain ISO string here on purpose (that's genuinely
// what arrives over the wire); see PostModel for the render-ready shape.
export interface Post {
	id: number;
	title: string;
	description: string;
	imageUrl: string;
	url: string;
	feedName: string;
	publishedAt: string;
	// ISO strings when set (matching Feed.disabledAt's convention), null
	// for the vast majority of posts which have never been saved.
	favoritedAt: string | null;
	readLaterAt: string | null;
}

// PostModel is what components actually render - fields already shaped
// for display. Built from a Post via toPostModel() (see postModel.ts).
// isFavorited/isReadLater are plain booleans, not the raw timestamps Post
// carries - components only ever need to know whether to show a badge, not
// when it was saved, so the null-check happens once here rather than in
// every component that renders one.
export interface PostModel {
	id: number;
	title: string;
	description: string;
	imageUrl: string;
	url: string;
	feedName: string;
	publishedAtDisplay: string;
	isFavorited: boolean;
	isReadLater: boolean;
}

// Feed matches exactly what the API client delivers for a feed - raw,
// unformatted. disabledAt is an ISO string when the feed has been removed
// (soft delete - see CLAUDE.md), null when active.
export interface Feed {
	id: number;
	name: string;
	url: string;
	disabledAt: string | null;
}

// FeedFailure identifies a feed that failed during a fetch, and why - lets
// the dashboard show which feed broke (e.g. an expired TLS certificate)
// instead of only a bare error count.
export interface FeedFailure {
	feedName: string;
	error: string;
}

// FetchSummary reports what an on-demand fetch (POST /api/fetch) did.
// newCount, not new - "new" is a reserved word in JS/TS, so it can't be
// used as a destructured binding name (`const { new } = summary` is a
// syntax error) even though it's fine as a plain property key. Renamed
// here rather than carrying that friction into every caller.
export interface FetchSummary {
	newCount: number;
	duplicates: number;
	errors: number;
	failedFeeds: FeedFailure[];
}
