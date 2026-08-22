import { API_URL } from '$env/static/private';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
	addFeed,
	deleteFeed,
	enableFeed,
	favoritePost,
	fetchNow,
	getFavoritedPosts,
	getFeeds,
	getPostsByDate,
	getPostsToday,
	getReadLaterPosts,
	markReadLater,
	unfavoritePost,
	unmarkReadLater,
	updateFeed
} from './api';

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
				fetched_at: '2026-07-10T09:05:00Z',
				favorited_at: null,
				read_later_at: null
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
					publishedAt: '2026-07-10T09:00:00Z',
					favoritedAt: null,
					readLaterAt: null
				}
			]);
		});

		it('maps favorited_at/read_later_at through when a post has been saved', async () => {
			// given: the API reports a post that's both favorited and read-later
			const wirePost = {
				id: 1,
				feed_id: 2,
				feed_name: 'Tech Blog',
				title: 'Go 2.0 Announced',
				url: 'https://example.com/go-2-0',
				content_summary: 'The Go team announces the next major version.',
				image_url: 'https://example.com/go-2-0.jpg',
				published_at: '2026-07-10T09:00:00Z',
				fetched_at: '2026-07-10T09:05:00Z',
				favorited_at: '2026-07-11T08:00:00Z',
				read_later_at: '2026-07-11T08:01:00Z'
			};
			vi.mocked(fetch).mockResolvedValue(mockResponse([wirePost]));

			// when: requesting posts for a specific date
			const posts = await getPostsByDate('2026-07-10');

			// then: both saved-state fields map through as their camelCase names
			expect(posts[0].favoritedAt).toBe('2026-07-11T08:00:00Z');
			expect(posts[0].readLaterAt).toBe('2026-07-11T08:01:00Z');
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
				fetched_at: '2026-07-13T11:20:00Z',
				favorited_at: null,
				read_later_at: null
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
					publishedAt: '2026-07-13T11:15:00Z',
					favoritedAt: null,
					readLaterAt: null
				}
			]);
		});
	});

	describe('getFavoritedPosts', () => {
		it('requests /api/posts/favorites and maps the wire post to a Post', async () => {
			// given: the API returns one favorited post
			const wirePost = {
				id: 4,
				feed_id: 5,
				feed_name: 'Cooking Blog',
				title: 'Perfect Sourdough Starter',
				url: 'https://example.com/sourdough',
				content_summary: 'A no-fuss guide to your first starter.',
				image_url: '',
				published_at: '2026-07-13T11:15:00Z',
				fetched_at: '2026-07-13T11:20:00Z',
				favorited_at: '2026-07-14T09:00:00Z',
				read_later_at: null
			};
			vi.mocked(fetch).mockResolvedValue(mockResponse([wirePost]));

			// when: requesting favorited posts
			const posts = await getFavoritedPosts();

			// then: fetch is called against /api/posts/favorites, and the post is mapped
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/posts/favorites`, undefined);
			expect(posts).toEqual([
				{
					id: 4,
					title: 'Perfect Sourdough Starter',
					description: 'A no-fuss guide to your first starter.',
					imageUrl: '',
					url: 'https://example.com/sourdough',
					feedName: 'Cooking Blog',
					publishedAt: '2026-07-13T11:15:00Z',
					favoritedAt: '2026-07-14T09:00:00Z',
					readLaterAt: null
				}
			]);
		});
	});

	describe('getReadLaterPosts', () => {
		it('requests /api/posts/read-later and maps the wire post to a Post', async () => {
			// given: the API returns one read-later post
			const wirePost = {
				id: 6,
				feed_id: 7,
				feed_name: 'Travel Blog',
				title: 'A Weekend in Ljubljana',
				url: 'https://example.com/ljubljana',
				content_summary: 'Small city, big charm.',
				image_url: '',
				published_at: '2026-07-13T11:15:00Z',
				fetched_at: '2026-07-13T11:20:00Z',
				favorited_at: null,
				read_later_at: '2026-07-14T09:01:00Z'
			};
			vi.mocked(fetch).mockResolvedValue(mockResponse([wirePost]));

			// when: requesting read-later posts
			const posts = await getReadLaterPosts();

			// then: fetch is called against /api/posts/read-later, and the post is mapped
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/posts/read-later`, undefined);
			expect(posts).toEqual([
				{
					id: 6,
					title: 'A Weekend in Ljubljana',
					description: 'Small city, big charm.',
					imageUrl: '',
					url: 'https://example.com/ljubljana',
					feedName: 'Travel Blog',
					publishedAt: '2026-07-13T11:15:00Z',
					favoritedAt: null,
					readLaterAt: '2026-07-14T09:01:00Z'
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

	describe('updateFeed', () => {
		it('puts the corrected name and url, and maps the updated feed', async () => {
			// given: the API accepts the correction and returns the updated feed
			const wireFeed = {
				id: 3,
				name: 'Corrected Blog',
				url: 'https://example.com/corrected-feed',
				disabled_at: null,
				created_at: '2026-07-13T00:00:00Z',
				updated_at: '2026-08-22T00:00:00Z'
			};
			vi.mocked(fetch).mockResolvedValue(mockResponse(wireFeed, 200));

			// when: correcting a feed's name/url
			const feed = await updateFeed(3, 'Corrected Blog', 'https://example.com/corrected-feed');

			// then: fetch is called with PUT against that feed's id, with the corrected body
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/feeds/3`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: 'Corrected Blog', url: 'https://example.com/corrected-feed' })
			});
			expect(feed).toEqual({
				id: 3,
				name: 'Corrected Blog',
				url: 'https://example.com/corrected-feed',
				disabledAt: null
			});
		});
	});

	describe('enableFeed', () => {
		it('sends a POST request and maps the re-enabled feed', async () => {
			// given: the API clears disabled_at and returns the updated feed
			const wireFeed = {
				id: 3,
				name: 'Re-enabled Blog',
				url: 'https://example.com/re-enabled-feed',
				disabled_at: null,
				created_at: '2026-07-13T00:00:00Z',
				updated_at: '2026-08-22T00:00:00Z'
			};
			vi.mocked(fetch).mockResolvedValue(mockResponse(wireFeed, 200));

			// when: enabling a disabled feed
			const feed = await enableFeed(3);

			// then: fetch is called with POST against that feed's enable path
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/feeds/3/enable`, { method: 'POST' });
			expect(feed).toEqual({
				id: 3,
				name: 'Re-enabled Blog',
				url: 'https://example.com/re-enabled-feed',
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

	describe('fetchNow', () => {
		it('sends a POST request and maps the wire summary, renaming new to newCount', async () => {
			// given: the API reports 3 new posts, 1 duplicate, 0 errors, no failures
			vi.mocked(fetch).mockResolvedValue(
				mockResponse({ new: 3, duplicates: 1, errors: 0, failed_feeds: [] })
			);

			// when: triggering an on-demand fetch
			const summary = await fetchNow();

			// then: fetch is called with POST, and the wire's "new" field maps to newCount
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/fetch`, { method: 'POST' });
			expect(summary).toEqual({ newCount: 3, duplicates: 1, errors: 0, failedFeeds: [] });
		});

		it('maps failed_feeds entries to camelCase feedName/error', async () => {
			// given: the API reports one failed feed with a real error message
			vi.mocked(fetch).mockResolvedValue(
				mockResponse({
					new: 0,
					duplicates: 0,
					errors: 1,
					failed_feeds: [
						{ feed_name: 'Sputnikmusic Staff Blog', error: 'tls: certificate has expired' }
					]
				})
			);

			// when: triggering an on-demand fetch
			const summary = await fetchNow();

			// then: the failed feed's name and error map to camelCase, unsanitized
			expect(summary.failedFeeds).toEqual([
				{ feedName: 'Sputnikmusic Staff Blog', error: 'tls: certificate has expired' }
			]);
		});
	});

	describe('favoritePost', () => {
		it('sends a POST request and maps the returned saved state', async () => {
			// given: the API favorites the post and returns its saved state
			vi.mocked(fetch).mockResolvedValue(
				mockResponse({ favorited_at: '2026-08-22T09:00:00Z', read_later_at: null })
			);

			// when: favoriting a post
			const state = await favoritePost(3);

			// then: fetch is called with POST against that post's favorite path
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/posts/3/favorite`, { method: 'POST' });
			expect(state).toEqual({ favoritedAt: '2026-08-22T09:00:00Z', readLaterAt: null });
		});
	});

	describe('unfavoritePost', () => {
		it('sends a DELETE request and maps the returned saved state', async () => {
			// given: the API unfavorites the post
			vi.mocked(fetch).mockResolvedValue(mockResponse({ favorited_at: null, read_later_at: null }));

			// when: unfavoriting a post
			const state = await unfavoritePost(3);

			// then: fetch is called with DELETE against that post's favorite path
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/posts/3/favorite`, { method: 'DELETE' });
			expect(state).toEqual({ favoritedAt: null, readLaterAt: null });
		});
	});

	describe('markReadLater', () => {
		it('sends a POST request and maps the returned saved state', async () => {
			// given: the API marks the post read-later
			vi.mocked(fetch).mockResolvedValue(
				mockResponse({ favorited_at: null, read_later_at: '2026-08-22T09:01:00Z' })
			);

			// when: marking a post read-later
			const state = await markReadLater(3);

			// then: fetch is called with POST against that post's read-later path
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/posts/3/read-later`, { method: 'POST' });
			expect(state).toEqual({ favoritedAt: null, readLaterAt: '2026-08-22T09:01:00Z' });
		});
	});

	describe('unmarkReadLater', () => {
		it('sends a DELETE request and maps the returned saved state', async () => {
			// given: the API unmarks the post
			vi.mocked(fetch).mockResolvedValue(mockResponse({ favorited_at: null, read_later_at: null }));

			// when: unmarking a post's read-later state
			const state = await unmarkReadLater(3);

			// then: fetch is called with DELETE against that post's read-later path
			expect(fetch).toHaveBeenCalledWith(`${API_URL}/api/posts/3/read-later`, {
				method: 'DELETE'
			});
			expect(state).toEqual({ favoritedAt: null, readLaterAt: null });
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
