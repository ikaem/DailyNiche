<!--
	One row in the dashboard's Active/Disabled Feeds lists - shared by both
	sections so the edit-toggle behavior below isn't duplicated between them.
	Disabled feeds get an inert "Enable" button instead of "Delete" (see the
	TODO below), but both kinds of feed can be edited - a disabled feed can
	still have a wrong name/url worth fixing.
-->
<script lang="ts">
	import { enhance } from '$app/forms';
	import type { Feed } from '$lib/types';

	let { feed }: { feed: Feed } = $props();

	let editing = $state(false);
</script>

{#if editing}
	<form
		method="POST"
		action="?/editFeed"
		class="edit-feed-form"
		use:enhance={() => {
			return async ({ result, update }) => {
				// Only close edit mode on success - stays open on a validation
				// failure (e.g. duplicate URL) so the fields aren't lost and the
				// user can immediately correct and resubmit.
				if (result.type === 'success') {
					editing = false;
				}
				await update();
			};
		}}
	>
		<input type="hidden" name="id" value={feed.id} />
		<div class="field">
			<label for="name-{feed.id}">Name</label>
			<input type="text" id="name-{feed.id}" name="name" value={feed.name} required />
		</div>
		<div class="field">
			<label for="url-{feed.id}">URL</label>
			<input type="url" id="url-{feed.id}" name="url" value={feed.url} required />
		</div>
		<div class="form-actions">
			<button type="button" class="cancel" onclick={() => (editing = false)}>Cancel</button>
			<button type="submit" class="primary">Save</button>
		</div>
	</form>
{:else}
	<div class="feed-row">
		<div class="feed-info">
			<span class="feed-name">{feed.name}</span>
			<span class="feed-url">{feed.url}</span>
		</div>
		<div class="feed-actions">
			<button type="button" class="edit" onclick={() => (editing = true)}>Edit</button>
			{#if feed.disabledAt === null}
				<form method="POST" action="?/deleteFeed" use:enhance>
					<input type="hidden" name="id" value={feed.id} />
					<button type="submit" class="delete">Delete</button>
				</form>
			{:else}
				<!-- Not wired to anything yet - repos.EnableFeed doesn't exist,
				     see the Task 7.4 TODO in CLAUDE.md. Looks and behaves like a
				     real button, just does nothing when clicked. -->
				<button type="button" class="enable">Enable</button>
			{/if}
		</div>
	</div>
{/if}

<style>
	.feed-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
		padding: 0.85rem 0;
		border-bottom: 1px solid var(--line);
	}
	.feed-row .feed-info {
		min-width: 0;
	}
	.feed-row .feed-name {
		font-weight: 600;
		font-size: 0.95rem;
	}
	.feed-row .feed-url {
		font-size: 0.8rem;
		color: var(--ink-soft);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		display: block;
	}
	.feed-actions {
		display: flex;
		flex-shrink: 0;
		gap: 0.5rem;
	}
	.feed-row button {
		flex-shrink: 0;
		background: none;
		border: 1px solid var(--line);
		border-radius: 6px;
		padding: 0.4rem 0.9rem;
		font-family: inherit;
		font-size: 0.82rem;
		font-weight: 500;
		cursor: pointer;
		color: var(--ink);
	}
	.feed-row button.delete {
		color: var(--accent);
		border-color: var(--accent-soft);
	}
	.feed-row button.delete:hover {
		background: var(--accent-soft);
	}
	.feed-row button.enable:hover {
		background: var(--paper);
	}
	.feed-row button.edit:hover {
		background: var(--card);
	}

	.edit-feed-form {
		border: 1px solid var(--line);
		border-radius: 8px;
		padding: 1.25rem 1.5rem;
		margin: 0.6rem 0;
		background: var(--card);
	}
	.field {
		margin-bottom: 1rem;
	}
	.field label {
		display: block;
		font-size: 0.85rem;
		font-weight: 500;
		margin-bottom: 0.35rem;
	}
	.field input {
		width: 100%;
		padding: 0.55rem 0.7rem;
		border: 1px solid var(--line);
		border-radius: 6px;
		font-family: inherit;
		font-size: 0.95rem;
		background: var(--paper);
		color: var(--ink);
	}
	.field input:focus {
		outline: none;
		border-color: var(--accent);
	}
	.form-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.6rem;
	}
	button.primary {
		background: var(--accent);
		color: #fff;
		border: none;
		border-radius: 6px;
		padding: 0.6rem 1.4rem;
		font-size: 0.9rem;
		font-weight: 600;
		cursor: pointer;
	}
	button.primary:hover {
		opacity: 0.92;
	}
	button.cancel {
		background: none;
		border: 1px solid var(--line);
		border-radius: 6px;
		padding: 0.6rem 1.4rem;
		font-family: inherit;
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--ink);
		cursor: pointer;
	}
	button.cancel:hover {
		background: var(--paper);
	}

	@media (max-width: 640px) {
		.feed-row {
			flex-wrap: wrap;
		}
		.feed-row .feed-info {
			flex: 1 1 100%;
		}
	}
</style>
