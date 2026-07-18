import { API_URL } from '$env/static/private';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { addFeed, deleteFeed, getFeeds, getPostsByDate, getPostsToday } from './api';

function mockResponse(body: unknown, status = 200): Response {
	return {
		ok: status >= 200 && status < 300,
		status,
		json: () => Promise.resolve(body)
	} as Response;
}

describe('api', () => {
	beforeEach(() => {
		vi.stubGlobal('fetch', vi.fn());
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	describe('getPostsByDate', () => {
		it('requests the given date and maps the wire post to a Post', async () => {
			// given: the API returns one post in its snake_case wire shape
			const wirePost = {
				id: 1,
				feed_id: 2,
				feed_name: 'Tech Blog',
				title: 'Go 2.0 Announced',
				url: 'https://example.com/go-2-0',
				content_summary: 'The Go team announces the next major version.',
				image_url: 'https://example.com/go-2-0.jpg',
				published_at: '2026-07-10T09:00:00Z',
				fetched_at: '2026-07-10T09:05:00Z'
			};
			vi.mocked(fetch).mockResolvedValue(mockResponse([wirePost]));

			// when: requesting posts for a specific date
			const posts = await getPostsByDate('2026-07-10');

			// then: fetch is called with the date query param, and the post is mapped
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/posts?date=2026-07-10`, undefined);
			expect(posts).toEqual([
				{
					id: 1,
					title: 'Go 2.0 Announced',
					description: 'The Go team announces the next major version.',
					imageUrl: 'https://example.com/go-2-0.jpg',
					url: 'https://example.com/go-2-0',
					feedName: 'Tech Blog',
					publishedAt: '2026-07-10T09:00:00Z'
				}
			]);
		});
	});

	describe('getPostsToday', () => {
		it('requests posts with no date param and maps the wire post to a Post', async () => {
			// given: the API returns one post in its snake_case wire shape
			const wirePost = {
				id: 4,
				feed_id: 5,
				feed_name: 'Cooking Blog',
				title: 'Perfect Sourdough Starter',
				url: 'https://example.com/sourdough',
				content_summary: 'A no-fuss guide to your first starter.',
				image_url: '',
				published_at: '2026-07-13T11:15:00Z',
				fetched_at: '2026-07-13T11:20:00Z'
			};
			vi.mocked(fetch).mockResolvedValue(mockResponse([wirePost]));

			// when: requesting today's posts
			const posts = await getPostsToday();

			// then: fetch is called against /api/posts with no query string, and the post is
			// mapped (including passing an empty image_url through as-is - the Go API is
			// responsible for ever substituting a placeholder, not this mapping layer)
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/posts`, undefined);
			expect(posts).toEqual([
				{
					id: 4,
					title: 'Perfect Sourdough Starter',
					description: 'A no-fuss guide to your first starter.',
					imageUrl: '',
					url: 'https://example.com/sourdough',
					feedName: 'Cooking Blog',
					publishedAt: '2026-07-13T11:15:00Z'
				}
			]);
		});
	});

	describe('getFeeds', () => {
		it('maps wire feeds to Feed, passing disabledAt through', async () => {
			// given: the API returns one active and one disabled feed
			const wireFeeds = [
				{
					id: 1,
					name: 'Tech Blog',
					url: 'https://example.com/feed',
					disabled_at: null,
					created_at: '2026-01-01T00:00:00Z',
					updated_at: '2026-01-01T00:00:00Z'
				},
				{
					id: 2,
					name: 'Old Blog',
					url: 'https://example.com/old-feed',
					disabled_at: '2026-05-01T00:00:00Z',
					created_at: '2026-01-01T00:00:00Z',
					updated_at: '2026-05-01T00:00:00Z'
				}
			];
			vi.mocked(fetch).mockResolvedValue(mockResponse(wireFeeds));

			// when: requesting all feeds
			const feeds = await getFeeds();

			// then: each feed is mapped to its camelCase shape
			expect(feeds).toEqual([
				{ id: 1, name: 'Tech Blog', url: 'https://example.com/feed', disabledAt: null },
				{
					id: 2,
					name: 'Old Blog',
					url: 'https://example.com/old-feed',
					disabledAt: '2026-05-01T00:00:00Z'
				}
			]);
		});
	});

	describe('addFeed', () => {
		it('posts the name and url, and maps the created feed', async () => {
			// given: the API creates and returns the new feed
			const wireFeed = {
				id: 3,
				name: 'New Blog',
				url: 'https://example.com/new-feed',
				disabled_at: null,
				created_at: '2026-07-13T00:00:00Z',
				updated_at: '2026-07-13T00:00:00Z'
			};
			vi.mocked(fetch).mockResolvedValue(mockResponse(wireFeed, 201));

			// when: adding a new feed
			const feed = await addFeed('New Blog', 'https://example.com/new-feed');

			// then: fetch is called with the right method, headers, and body
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/feeds`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: 'New Blog', url: 'https://example.com/new-feed' })
			});
			expect(feed).toEqual({
				id: 3,
				name: 'New Blog',
				url: 'https://example.com/new-feed',
				disabledAt: null
			});
		});
	});

	describe('deleteFeed', () => {
		it('sends a DELETE request for the given id', async () => {
			// given: the API accepts the delete with no response body
			vi.mocked(fetch).mockResolvedValue(mockResponse(null, 204));

			// when: deleting a feed
			await deleteFeed(3);

			// then: fetch is called with DELETE against that feed's id
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/feeds/3`, { method: 'DELETE' });
		});
	});

	describe('error handling', () => {
		it('throws an ApiError carrying the status when the response is not ok', async () => {
			// given: the API rejects the request (e.g. duplicate feed URL)
			vi.mocked(fetch).mockResolvedValue(mockResponse(null, 409));

			// when: adding a feed that the API rejects
			// then: the promise rejects with an ApiError carrying that status
			await expect(addFeed('Dup', 'https://example.com/dup')).rejects.toMatchObject({
				name: 'ApiError',
				status: 409
			});
		});

		it('uses the Go API\'s {"error": "..."} body as the message when present', async () => {
			// given: the API rejects with its real JSON error body
			vi.mocked(fetch).mockResolvedValue(
				mockResponse({ error: 'a feed with this url already exists' }, 409)
			);

			// when: adding a feed that the API rejects
			// then: the ApiError's message is the API's actual reason, not a generic one
			await expect(addFeed('Dup', 'https://example.com/dup')).rejects.toMatchObject({
				message: 'a feed with this url already exists',
				status: 409
			});
		});

		it('falls back to a generic message when the error body is not valid JSON', async () => {
			// given: the response body isn't JSON at all (e.g. an intermediary
			// returned something other than the Go API's own error response)
			const notJsonResponse = {
				ok: false,
				status: 502,
				json: () => Promise.reject(new SyntaxError('Unexpected token'))
			} as Response;
			vi.mocked(fetch).mockResolvedValue(notJsonResponse);

			// when: adding a feed and the response can't be parsed as JSON
			// then: it falls back to the generic status-based message, not a thrown SyntaxError
			await expect(addFeed('Dup', 'https://example.com/dup')).rejects.toMatchObject({
				message: 'API request to /api/feeds failed with status 502',
				status: 502
			});
		});
	});
});
