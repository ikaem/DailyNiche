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
}

// PostModel is what components actually render - fields already shaped
// for display. Built from a Post via toPostModel() (see postModel.ts).
export interface PostModel {
	id: number;
	title: string;
	description: string;
	imageUrl: string;
	url: string;
	feedName: string;
	publishedAtDisplay: string;
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
