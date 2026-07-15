import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Feed } from '$lib/types';
import { ApiError } from '$lib/server/api';

// See src/routes/page.server.spec.ts for why vi.hoisted + vi.mock are
// paired like this. ApiError is spread through from the real module (via
// vi.importActual) rather than faked, so `instanceof ApiError` inside the
// action still works correctly against errors thrown in these tests.
const { getFeeds, addFeed, deleteFeed } = vi.hoisted(() => ({
	getFeeds: vi.fn(),
	addFeed: vi.fn(),
	deleteFeed: vi.fn()
}));
vi.mock('$lib/server/api', async () => {
	const actual = await vi.importActual<typeof import('$lib/server/api')>('$lib/server/api');
	return { ...actual, getFeeds, addFeed, deleteFeed };
});

import { actions, load } from './+page.server';

beforeEach(() => {
	getFeeds.mockReset();
	addFeed.mockReset();
	deleteFeed.mockReset();
});

function formDataRequest(fields: Record<string, string>): Request {
	const formData = new FormData();
	for (const [key, value] of Object.entries(fields)) {
		formData.append(key, value);
	}
	return new Request('http://localhost/dashboard', { method: 'POST', body: formData });
}

describe('dashboard +page.server load', () => {
	it('returns feeds and no error on success', async () => {
		// given: getFeeds resolves with an active and a disabled feed
		const feeds: Feed[] = [
			{ id: 1, name: 'Tech Blog', url: 'https://example.com/tech/feed.xml', disabledAt: null },
			{
				id: 2,
				name: 'Old Blog',
				url: 'https://example.com/old/feed.xml',
				disabledAt: '2026-05-01T00:00:00Z'
			}
		];
		getFeeds.mockResolvedValue(feeds);

		// when: the load function runs
		const result = await load({} as Parameters<typeof load>[0]);

		// then: it returns the feeds with no error
		expect(result).toEqual({ feeds, error: null });
	});

	it('returns an empty list and an error message when the fetch fails', async () => {
		// given: getFeeds rejects
		getFeeds.mockRejectedValue(new Error('network down'));

		// when: the load function runs
		const result = await load({} as Parameters<typeof load>[0]);

		// then: it returns an empty list with the error message, not a thrown error
		expect(result).toEqual({ feeds: [], error: 'network down' });
	});
});

describe('dashboard actions.addFeed', () => {
	it('trims and forwards the form fields, returning nothing on success', async () => {
		// given: addFeed resolves, and the submitted fields have stray whitespace
		addFeed.mockResolvedValue({
			id: 3,
			name: 'New Blog',
			url: 'https://example.com/new-feed',
			disabledAt: null
		});
		const request = formDataRequest({
			name: '  New Blog  ',
			url: '  https://example.com/new-feed  '
		});

		// when: the action runs
		const result = await actions.addFeed({ request } as Parameters<typeof actions.addFeed>[0]);

		// then: addFeed is called with the trimmed name/url, and nothing is returned
		expect(addFeed).toHaveBeenCalledWith('New Blog', 'https://example.com/new-feed');
		expect(result).toBeUndefined();
	});

	it('fails with 400 and does not call addFeed when name is missing', async () => {
		// given: a submission with no name
		const request = formDataRequest({ name: '', url: 'https://example.com/new-feed' });

		// when: the action runs
		const result = await actions.addFeed({ request } as Parameters<typeof actions.addFeed>[0]);

		// then: it fails validation before ever calling addFeed
		expect(result).toEqual({ status: 400, data: { message: 'name is required' } });
		expect(addFeed).not.toHaveBeenCalled();
	});

	it('fails with 400 and does not call addFeed when url is missing', async () => {
		// given: a submission with no url
		const request = formDataRequest({ name: 'New Blog', url: '' });

		// when: the action runs
		const result = await actions.addFeed({ request } as Parameters<typeof actions.addFeed>[0]);

		// then: it fails validation before ever calling addFeed
		expect(result).toEqual({ status: 400, data: { message: 'url is required' } });
		expect(addFeed).not.toHaveBeenCalled();
	});

	it('fails with 400 and does not call addFeed when url is not a valid absolute URL', async () => {
		// given: a submission with a malformed url
		const request = formDataRequest({ name: 'New Blog', url: 'not-a-url' });

		// when: the action runs
		const result = await actions.addFeed({ request } as Parameters<typeof actions.addFeed>[0]);

		// then: it fails validation before ever calling addFeed
		expect(result).toEqual({ status: 400, data: { message: 'url must be a valid absolute URL' } });
		expect(addFeed).not.toHaveBeenCalled();
	});

	it('fails with the ApiError status and message when the Go API rejects the feed', async () => {
		// given: addFeed rejects with an ApiError (e.g. duplicate URL, 409)
		addFeed.mockRejectedValue(new ApiError('a feed with this url already exists', 409));
		const request = formDataRequest({ name: 'Dup', url: 'https://example.com/dup' });

		// when: the action runs
		const result = await actions.addFeed({ request } as Parameters<typeof actions.addFeed>[0]);

		// then: it returns the same status and message as the ApiError
		expect(result).toEqual({
			status: 409,
			data: { message: 'a feed with this url already exists' }
		});
	});

	it('fails with 500 when addFeed throws a non-ApiError error', async () => {
		// given: addFeed rejects with an unexpected error
		addFeed.mockRejectedValue(new Error('connection reset'));
		const request = formDataRequest({ name: 'New Blog', url: 'https://example.com/new-feed' });

		// when: the action runs
		const result = await actions.addFeed({ request } as Parameters<typeof actions.addFeed>[0]);

		// then: it falls back to a generic 500 message
		expect(result).toEqual({ status: 500, data: { message: 'Failed to add feed' } });
	});
});

describe('dashboard actions.deleteFeed', () => {
	it('forwards the numeric id, returning nothing on success', async () => {
		// given: deleteFeed resolves
		deleteFeed.mockResolvedValue(undefined);
		const request = formDataRequest({ id: '3' });

		// when: the action runs
		const result = await actions.deleteFeed({ request } as Parameters<
			typeof actions.deleteFeed
		>[0]);

		// then: deleteFeed is called with the id as a number, and nothing is returned
		expect(deleteFeed).toHaveBeenCalledWith(3);
		expect(result).toBeUndefined();
	});

	it('fails with 400 and does not call deleteFeed when id is missing', async () => {
		// given: a submission with no id
		const request = formDataRequest({});

		// when: the action runs
		const result = await actions.deleteFeed({ request } as Parameters<
			typeof actions.deleteFeed
		>[0]);

		// then: it fails validation before ever calling deleteFeed
		expect(result).toEqual({ status: 400, data: { message: 'a valid feed id is required' } });
		expect(deleteFeed).not.toHaveBeenCalled();
	});

	it('fails with 400 and does not call deleteFeed when id is not numeric', async () => {
		// given: a submission with a non-numeric id
		const request = formDataRequest({ id: 'abc' });

		// when: the action runs
		const result = await actions.deleteFeed({ request } as Parameters<
			typeof actions.deleteFeed
		>[0]);

		// then: it fails validation before ever calling deleteFeed
		expect(result).toEqual({ status: 400, data: { message: 'a valid feed id is required' } });
		expect(deleteFeed).not.toHaveBeenCalled();
	});

	it('fails with the ApiError status and message when the Go API rejects the deletion', async () => {
		// given: deleteFeed rejects with an ApiError (e.g. not found, 404)
		deleteFeed.mockRejectedValue(new ApiError('feed not found', 404));
		const request = formDataRequest({ id: '99' });

		// when: the action runs
		const result = await actions.deleteFeed({ request } as Parameters<
			typeof actions.deleteFeed
		>[0]);

		// then: it returns the same status and message as the ApiError
		expect(result).toEqual({ status: 404, data: { message: 'feed not found' } });
	});

	it('fails with 500 when deleteFeed throws a non-ApiError error', async () => {
		// given: deleteFeed rejects with an unexpected error
		deleteFeed.mockRejectedValue(new Error('connection reset'));
		const request = formDataRequest({ id: '3' });

		// when: the action runs
		const result = await actions.deleteFeed({ request } as Parameters<
			typeof actions.deleteFeed
		>[0]);

		// then: it falls back to a generic 500 message
		expect(result).toEqual({ status: 500, data: { message: 'Failed to delete feed' } });
	});
});
