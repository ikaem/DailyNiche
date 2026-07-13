import { PUBLIC_API_URL } from '$env/static/public';
import type { Feed, Post } from './types';

// ApiError carries the HTTP status alongside the message, so callers (e.g.
// FeedManager, Task 7.4) can branch on it - a 409 duplicate-URL response
// needs different UI feedback than a 500.
export class ApiError extends Error {
	status: number;

	constructor(message: string, status: number) {
		super(message);
		this.name = 'ApiError';
		this.status = status;
	}
}

// Wire shapes match the Go API's JSON exactly (snake_case, as encoded by
// PostResponse/FeedResponse) - kept private to this module, never exposed
// to callers, who only ever see the camelCase Post/Feed types.
interface PostWire {
	id: number;
	feed_id: number;
	feed_name: string;
	title: string;
	url: string;
	content_summary: string;
	published_at: string;
	fetched_at: string;
}

interface FeedWire {
	id: number;
	name: string;
	url: string;
	disabled_at: string | null;
	created_at: string;
	updated_at: string;
}

// TODO: imageUrl is always empty until Task 2.3 (post images pipeline) adds
// image_url to the API response - see CLAUDE.md.
function toPost(wire: PostWire): Post {
	return {
		id: wire.id,
		title: wire.title,
		description: wire.content_summary,
		imageUrl: '',
		url: wire.url,
		feedName: wire.feed_name,
		publishedAt: wire.published_at
	};
}

function toFeed(wire: FeedWire): Feed {
	return {
		id: wire.id,
		name: wire.name,
		url: wire.url,
		disabledAt: wire.disabled_at
	};
}

async function apiFetch(path: string, init?: RequestInit): Promise<Response> {
	const res = await fetch(`${PUBLIC_API_URL}${path}`, init);
	if (!res.ok) {
		throw new ApiError(`API request to ${path} failed with status ${res.status}`, res.status);
	}
	return res;
}

async function apiFetchJson<T>(path: string, init?: RequestInit): Promise<T> {
	const res = await apiFetch(path, init);
	return res.json();
}

export async function getPostsByDate(date: string): Promise<Post[]> {
	const wire = await apiFetchJson<PostWire[]>(`/api/posts?date=${date}`);
	return wire.map(toPost);
}

export async function getPostsToday(): Promise<Post[]> {
	const wire = await apiFetchJson<PostWire[]>('/api/posts');
	return wire.map(toPost);
}

export async function getFeeds(): Promise<Feed[]> {
	const wire = await apiFetchJson<FeedWire[]>('/api/feeds');
	return wire.map(toFeed);
}

export async function addFeed(name: string, url: string): Promise<Feed> {
	const wire = await apiFetchJson<FeedWire>('/api/feeds', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ name, url })
	});
	return toFeed(wire);
}

export async function deleteFeed(id: number): Promise<void> {
	await apiFetch(`/api/feeds/${id}`, { method: 'DELETE' });
}
