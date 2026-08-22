import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Post } from '$lib/types';

// See src/routes/page.server.spec.ts for why vi.hoisted + vi.mock are
// paired like this, and why ApiError is spread through via vi.importActual
// rather than faked.
const {
	getFavoritedPosts,
	getReadLaterPosts,
	favoritePost,
	unfavoritePost,
	markReadLater,
	unmarkReadLater
} = vi.hoisted(() => ({
	getFavoritedPosts: vi.fn(),
	getReadLaterPosts: vi.fn(),
	favoritePost: vi.fn(),
	unfavoritePost: vi.fn(),
	markReadLater: vi.fn(),
	unmarkReadLater: vi.fn()
}));
vi.mock('$lib/server/api', async () => {
	const actual = await vi.importActual<typeof import('$lib/server/api')>('$lib/server/api');
	return {
		...actual,
		getFavoritedPosts,
		getReadLaterPosts,
		favoritePost,
		unfavoritePost,
		markReadLater,
		unmarkReadLater
	};
});

import { ApiError } from '$lib/server/api';
import { actions, load } from './+page.server';

function formDataRequest(fields: Record<string, string>): Request {
	const formData = new FormData();
	for (const [key, value] of Object.entries(fields)) {
		formData.append(key, value);
	}
	return new Request('http://localhost/saved', { method: 'POST', body: formData });
}

beforeEach(() => {
	vi.clearAllMocks();
});

describe('+page.server load', () => {
	it('returns both lists and no error on success', async () => {
		// given: both lists resolve with a post each
		const favorited: Post[] = [
			{
				id: 1,
				title: 'Go 2.0 Announced',
				description: 'The Go team announces the next major version.',
				imageUrl: '',
				url: 'https://example.com/go-2-0',
				feedName: 'Tech Blog',
				publishedAt: '2026-07-10T09:00:00Z',
				favoritedAt: '2026-07-15T09:00:00Z',
				readLaterAt: null
			}
		];
		const readLater: Post[] = [
			{
				id: 2,
				title: 'Why SQLite Is Enough',
				description: 'A case for boring, reliable databases.',
				imageUrl: '',
				url: 'https://example.com/sqlite',
				feedName: 'Tech Blog',
				publishedAt: '2026-07-09T09:00:00Z',
				favoritedAt: null,
				readLaterAt: '2026-07-15T09:01:00Z'
			}
		];
		getFavoritedPosts.mockResolvedValue(favorited);
		getReadLaterPosts.mockResolvedValue(readLater);

		// when: the load function runs
		const result = await load({} as Parameters<typeof load>[0]);

		// then: it returns both lists with no error
		expect(result).toEqual({ favoritedPosts: favorited, readLaterPosts: readLater, error: null });
	});

	it('returns two empty lists and an error message when either fetch fails', async () => {
		// given: getReadLaterPosts rejects
		getFavoritedPosts.mockResolvedValue([]);
		getReadLaterPosts.mockRejectedValue(new Error('network down'));

		// when: the load function runs
		const result = await load({} as Parameters<typeof load>[0]);

		// then: it returns empty lists with the error message, not a thrown error
		expect(result).toEqual({
			favoritedPosts: [],
			readLaterPosts: [],
			error: 'network down'
		});
	});
});

describe('+page.server actions.unfavoritePost', () => {
	it('forwards the numeric id, returning nothing on success', async () => {
		// given: unfavoritePost resolves
		unfavoritePost.mockResolvedValue({ favoritedAt: null, readLaterAt: null });
		const request = formDataRequest({ id: '3' });

		// when: the action runs
		const result = await actions.unfavoritePost({ request } as Parameters<
			typeof actions.unfavoritePost
		>[0]);

		// then: unfavoritePost is called with the id as a number, and nothing is returned
		expect(unfavoritePost).toHaveBeenCalledWith(3);
		expect(result).toBeUndefined();
	});

	it('fails with the ApiError status and message when the Go API rejects it', async () => {
		// given: unfavoritePost rejects with an ApiError (e.g. not found, 404)
		unfavoritePost.mockRejectedValue(new ApiError('post not found', 404));
		const request = formDataRequest({ id: '99' });

		// when: the action runs
		const result = await actions.unfavoritePost({ request } as Parameters<
			typeof actions.unfavoritePost
		>[0]);

		// then: it returns the same status and message as the ApiError
		expect(result).toEqual({ status: 404, data: { message: 'post not found' } });
	});
});

describe('+page.server actions.unmarkReadLater', () => {
	it('forwards the numeric id, returning nothing on success', async () => {
		// given: unmarkReadLater resolves
		unmarkReadLater.mockResolvedValue({ favoritedAt: null, readLaterAt: null });
		const request = formDataRequest({ id: '3' });

		// when: the action runs
		const result = await actions.unmarkReadLater({ request } as Parameters<
			typeof actions.unmarkReadLater
		>[0]);

		// then: unmarkReadLater is called with the id as a number, and nothing is returned
		expect(unmarkReadLater).toHaveBeenCalledWith(3);
		expect(result).toBeUndefined();
	});

	it('fails with 400 and does not call unmarkReadLater when id is missing', async () => {
		// given: a submission with no id
		const request = formDataRequest({});

		// when: the action runs
		const result = await actions.unmarkReadLater({ request } as Parameters<
			typeof actions.unmarkReadLater
		>[0]);

		// then: it fails validation before ever calling unmarkReadLater
		expect(result).toEqual({ status: 400, data: { message: 'a valid post id is required' } });
		expect(unmarkReadLater).not.toHaveBeenCalled();
	});
});
