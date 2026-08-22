import { fail } from '@sveltejs/kit';
import { ApiError, favoritePost, markReadLater, unfavoritePost, unmarkReadLater } from './api';

// The four favorite/read-later toggle actions, shared by every route whose
// posts can be saved from (the home page, and the Saved page) - each
// route's own +page.server.ts still needs its own `export const actions`
// (SvelteKit requires that per-route), but assigns these instead of
// duplicating the same four function bodies.
//
// Typed against just { request: Request }, not a route's own generated
// Action type from ./$types - this module is shared across routes with
// different route-specific types, and these actions only ever read
// request.formData() anyway.

function parsePostId(formData: FormData): number | null {
	const raw = String(formData.get('id') ?? '');
	const id = Number(raw);
	if (!raw || Number.isNaN(id)) {
		return null;
	}
	return id;
}

export async function favoritePostAction({ request }: { request: Request }) {
	const id = parsePostId(await request.formData());
	if (id === null) {
		return fail(400, { message: 'a valid post id is required' });
	}
	try {
		await favoritePost(id);
	} catch (err) {
		if (err instanceof ApiError) {
			return fail(err.status, { message: err.message });
		}
		return fail(500, { message: 'Failed to favorite post' });
	}
}

export async function unfavoritePostAction({ request }: { request: Request }) {
	const id = parsePostId(await request.formData());
	if (id === null) {
		return fail(400, { message: 'a valid post id is required' });
	}
	try {
		await unfavoritePost(id);
	} catch (err) {
		if (err instanceof ApiError) {
			return fail(err.status, { message: err.message });
		}
		return fail(500, { message: 'Failed to unfavorite post' });
	}
}

export async function markReadLaterAction({ request }: { request: Request }) {
	const id = parsePostId(await request.formData());
	if (id === null) {
		return fail(400, { message: 'a valid post id is required' });
	}
	try {
		await markReadLater(id);
	} catch (err) {
		if (err instanceof ApiError) {
			return fail(err.status, { message: err.message });
		}
		return fail(500, { message: 'Failed to mark post read later' });
	}
}

export async function unmarkReadLaterAction({ request }: { request: Request }) {
	const id = parsePostId(await request.formData());
	if (id === null) {
		return fail(400, { message: 'a valid post id is required' });
	}
	try {
		await unmarkReadLater(id);
	} catch (err) {
		if (err instanceof ApiError) {
			return fail(err.status, { message: err.message });
		}
		return fail(500, { message: 'Failed to unmark post read later' });
	}
}
