import { describe, expect, it } from 'vitest';
import { toPostModel } from './postModel';
import type { Post } from './types';

describe('toPostModel', () => {
	it('formats publishedAt into a display date, passing other fields through unchanged', () => {
		// given: a raw Post with an ISO publishedAt
		const post: Post = {
			id: 1,
			title: 'Go 2.0 Announced',
			description: 'The Go team announces the next major version.',
			imageUrl: 'https://example.com/image.jpg',
			url: 'https://example.com/post',
			feedName: 'Tech Blog',
			publishedAt: '2026-07-08T17:45:56.319884647Z'
		};

		// when: we map it to a PostModel
		const model = toPostModel(post);

		// then: publishedAtDisplay is formatted, everything else passes through
		expect(model.publishedAtDisplay).toBe('8. srp 2026.');
		expect(model.id).toBe(post.id);
		expect(model.title).toBe(post.title);
		expect(model.description).toBe(post.description);
		expect(model.imageUrl).toBe(post.imageUrl);
		expect(model.url).toBe(post.url);
		expect(model.feedName).toBe(post.feedName);
	});
});
