import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Post } from '$lib/types';

// vi.mock's factory is hoisted above this file's imports, so it can't
// reference a normal variable declared further down (it wouldn't exist yet
// when the factory runs). vi.hoisted's own callback is hoisted too, and
// runs first, so its return value IS available inside vi.mock's factory.
// Together: every import of these functions from $lib/server/api -
// including the ones inside +page.server.ts below - resolves to these same
// shared fakes. ApiError is spread through from the real module (via
// vi.importActual) rather than faked, so `instanceof ApiError` inside the
// actions still works correctly against errors thrown in these tests - see
// the dashboard's own page.server.spec.ts for the same pattern.
const {
	getPostsToday,
	getPostsByDate,
	favoritePost,
	unfavoritePost,
	markReadLater,
	unmarkReadLater
} = vi.hoisted(() => ({
	getPostsToday: vi.fn(),
	getPostsByDate: vi.fn(),
	favoritePost: vi.fn(),
	unfavoritePost: vi.fn(),
	markReadLater: vi.fn(),
	unmarkReadLater: vi.fn()
}));
vi.mock('$lib/server/api', async () => {
	const actual = await vi.importActual<typeof import('$lib/server/api')>('$lib/server/api');
	return {
		...actual,
		getPostsToday,
		getPostsByDate,
		favoritePost,
		unfavoritePost,
		markReadLater,
		unmarkReadLater
	};
});

import { ApiError } from '$lib/server/api';
import { actions, load } from './+page.server';

// Builds a fake LoadEvent carrying just the "url" property load() actually
// reads - everything else on the real LoadEvent is irrelevant here.
function loadEvent(path: string): Parameters<typeof load>[0] {
	return { url: new URL(`http://localhost${path}`) } as Parameters<typeof load>[0];
}

function formDataRequest(fields: Record<string, string>): Request {
	const formData = new FormData();
	for (const [key, value] of Object.entries(fields)) {
		formData.append(key, value);
	}
	return new Request('http://localhost/', { method: 'POST', body: formData });
}

// File-level, not nested in one describe block - all six mocked functions
// are shared (via vi.hoisted) across every describe block below, so without
// this, call counts from one action's tests leak into another's.
beforeEach(() => {
	vi.clearAllMocks();
});

describe('+page.server load', () => {
	beforeEach(() => {
		// Freezes "today" so the date-less path resolves to a known, exact
		// date instead of whatever day the test happens to run on.
		vi.setSystemTime(new Date('2026-07-15T12:00:00Z'));
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('returns posts and no error on success', async () => {
		// given: getPostsToday resolves with one post
		const posts: Post[] = [
			{
				id: 1,
				title: 'Go 2.0 Announced',
				description: 'The Go team announces the next major version.',
				imageUrl: '',
				url: 'https://example.com/go-2-0-announced',
				feedName: 'Tech Blog',
				publishedAt: '2026-07-10T09:00:00Z',
				favoritedAt: null,
				readLaterAt: null
			}
		];
		getPostsToday.mockResolvedValue(posts);

		// when: the load function runs with no ?date param
		const result = await load(loadEvent('/'));

		// then: it returns the posts, no error, and today's date
		expect(result).toEqual({ posts, error: null, date: '2026-07-15' });
		expect(getPostsByDate).not.toHaveBeenCalled();
	});

	it('returns an empty list and an error message when the fetch fails', async () => {
		// given: getPostsToday rejects
		getPostsToday.mockRejectedValue(new Error('network down'));

		// when: the load function runs with no ?date param
		const result = await load(loadEvent('/'));

		// then: it returns an empty list with the error message, not a thrown error
		expect(result).toEqual({ posts: [], error: 'network down', date: '2026-07-15' });
	});

	it('fetches a specific date when ?date is present, instead of today', async () => {
		// given: getPostsByDate resolves with one post for a specific date
		const posts: Post[] = [
			{
				id: 2,
				title: 'Archived Post',
				description: 'From a previous day.',
				imageUrl: '',
				url: 'https://example.com/archived-post',
				feedName: 'Tech Blog',
				publishedAt: '2026-07-09T09:00:00Z',
				favoritedAt: null,
				readLaterAt: null
			}
		];
		getPostsByDate.mockResolvedValue(posts);

		// when: the load function runs with ?date=2026-07-09
		const result = await load(loadEvent('/?date=2026-07-09'));

		// then: it calls getPostsByDate with that exact date, not getPostsToday
		expect(getPostsByDate).toHaveBeenCalledWith('2026-07-09');
		expect(getPostsToday).not.toHaveBeenCalled();
		expect(result).toEqual({ posts, error: null, date: '2026-07-09' });
	});

	it('returns an empty list and an error message when fetching a specific date fails', async () => {
		// given: getPostsByDate rejects
		getPostsByDate.mockRejectedValue(new Error('network down'));

		// when: the load function runs with ?date=2026-07-09
		const result = await load(loadEvent('/?date=2026-07-09'));

		// then: it returns an empty list with the error message and the requested date
		expect(result).toEqual({ posts: [], error: 'network down', date: '2026-07-09' });
	});
});

describe('+page.server actions.favoritePost', () => {
	it('forwards the numeric id, returning nothing on success', async () => {
		// given: favoritePost resolves
		favoritePost.mockResolvedValue({ favoritedAt: '2026-07-15T09:00:00Z', readLaterAt: null });
		const request = formDataRequest({ id: '3' });

		// when: the action runs
		const result = await actions.favoritePost({ request } as Parameters<
			typeof actions.favoritePost
		>[0]);

		// then: favoritePost is called with the id as a number, and nothing is returned
		expect(favoritePost).toHaveBeenCalledWith(3);
		expect(result).toBeUndefined();
	});

	it('fails with 400 and does not call favoritePost when id is missing', async () => {
		// given: a submission with no id
		const request = formDataRequest({});

		// when: the action runs
		const result = await actions.favoritePost({ request } as Parameters<
			typeof actions.favoritePost
		>[0]);

		// then: it fails validation before ever calling favoritePost
		expect(result).toEqual({ status: 400, data: { message: 'a valid post id is required' } });
		expect(favoritePost).not.toHaveBeenCalled();
	});

	it('fails with the ApiError status and message when the Go API rejects it', async () => {
		// given: favoritePost rejects with an ApiError (e.g. not found, 404)
		favoritePost.mockRejectedValue(new ApiError('post not found', 404));
		const request = formDataRequest({ id: '99' });

		// when: the action runs
		const result = await actions.favoritePost({ request } as Parameters<
			typeof actions.favoritePost
		>[0]);

		// then: it returns the same status and message as the ApiError
		expect(result).toEqual({ status: 404, data: { message: 'post not found' } });
	});

	it('fails with 500 when favoritePost throws a non-ApiError error', async () => {
		// given: favoritePost rejects with an unexpected error
		favoritePost.mockRejectedValue(new Error('connection reset'));
		const request = formDataRequest({ id: '3' });

		// when: the action runs
		const result = await actions.favoritePost({ request } as Parameters<
			typeof actions.favoritePost
		>[0]);

		// then: it falls back to a generic 500 message
		expect(result).toEqual({ status: 500, data: { message: 'Failed to favorite post' } });
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

	it('fails with 400 and does not call unfavoritePost when id is not numeric', async () => {
		// given: a submission with a non-numeric id
		const request = formDataRequest({ id: 'abc' });

		// when: the action runs
		const result = await actions.unfavoritePost({ request } as Parameters<
			typeof actions.unfavoritePost
		>[0]);

		// then: it fails validation before ever calling unfavoritePost
		expect(result).toEqual({ status: 400, data: { message: 'a valid post id is required' } });
		expect(unfavoritePost).not.toHaveBeenCalled();
	});

	it('fails with 500 when unfavoritePost throws a non-ApiError error', async () => {
		// given: unfavoritePost rejects with an unexpected error
		unfavoritePost.mockRejectedValue(new Error('connection reset'));
		const request = formDataRequest({ id: '3' });

		// when: the action runs
		const result = await actions.unfavoritePost({ request } as Parameters<
			typeof actions.unfavoritePost
		>[0]);

		// then: it falls back to a generic 500 message
		expect(result).toEqual({ status: 500, data: { message: 'Failed to unfavorite post' } });
	});
});

describe('+page.server actions.markReadLater', () => {
	it('forwards the numeric id, returning nothing on success', async () => {
		// given: markReadLater resolves
		markReadLater.mockResolvedValue({ favoritedAt: null, readLaterAt: '2026-07-15T09:00:00Z' });
		const request = formDataRequest({ id: '3' });

		// when: the action runs
		const result = await actions.markReadLater({ request } as Parameters<
			typeof actions.markReadLater
		>[0]);

		// then: markReadLater is called with the id as a number, and nothing is returned
		expect(markReadLater).toHaveBeenCalledWith(3);
		expect(result).toBeUndefined();
	});

	it('fails with 400 and does not call markReadLater when id is missing', async () => {
		// given: a submission with no id
		const request = formDataRequest({});

		// when: the action runs
		const result = await actions.markReadLater({ request } as Parameters<
			typeof actions.markReadLater
		>[0]);

		// then: it fails validation before ever calling markReadLater
		expect(result).toEqual({ status: 400, data: { message: 'a valid post id is required' } });
		expect(markReadLater).not.toHaveBeenCalled();
	});

	it('fails with the ApiError status and message when the Go API rejects it', async () => {
		// given: markReadLater rejects with an ApiError (e.g. not found, 404)
		markReadLater.mockRejectedValue(new ApiError('post not found', 404));
		const request = formDataRequest({ id: '99' });

		// when: the action runs
		const result = await actions.markReadLater({ request } as Parameters<
			typeof actions.markReadLater
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

	it('fails with 400 and does not call unmarkReadLater when id is not numeric', async () => {
		// given: a submission with a non-numeric id
		const request = formDataRequest({ id: 'abc' });

		// when: the action runs
		const result = await actions.unmarkReadLater({ request } as Parameters<
			typeof actions.unmarkReadLater
		>[0]);

		// then: it fails validation before ever calling unmarkReadLater
		expect(result).toEqual({ status: 400, data: { message: 'a valid post id is required' } });
		expect(unmarkReadLater).not.toHaveBeenCalled();
	});

	it('fails with 500 when unmarkReadLater throws a non-ApiError error', async () => {
		// given: unmarkReadLater rejects with an unexpected error
		unmarkReadLater.mockRejectedValue(new Error('connection reset'));
		const request = formDataRequest({ id: '3' });

		// when: the action runs
		const result = await actions.unmarkReadLater({ request } as Parameters<
			typeof actions.unmarkReadLater
		>[0]);

		// then: it falls back to a generic 500 message
		expect(result).toEqual({ status: 500, data: { message: 'Failed to unmark post read later' } });
	});
});
