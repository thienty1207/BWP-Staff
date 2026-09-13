<svelte:options runes={true} />

<script lang="ts">
	import { goto } from '$app/navigation';
	import { getDepartments, getLocations } from '$lib/client/lookups/api';
	import { LookupApiError, type LookupDepartment, type LookupLocation } from '$lib/client/lookups/model';
	import { createTicket } from '$lib/client/tickets/api';
	import { TicketApiError, type CreateTicketErrorCode, type CreateTicketRequest } from '$lib/client/tickets/model';

	type Props = {
	open: boolean;
	disabled?: boolean;
	onClose: () => void;
	onCreated: () => Promise<void> | void;
};

	let { open, disabled = false, onClose, onCreated }: Props = $props();

	let departments: LookupDepartment[] = $state([]);
	let locations: LookupLocation[] = $state([]);
	let lookupLoading = $state(false);
	let lookupError = $state('');
	let lookupLoadStarted = $state(false);
	let submitting = $state(false);
	let validationError = $state('');
	let submitError = $state('');

	let departmentID = $state('');
	let locationID = $state('');
	let title = $state('');
	let description = $state('');
	let priority = $state(false);
	let dueAt = $state('');

	$effect(() => {
		if (open && !lookupLoadStarted) {
			lookupLoadStarted = true;
			void loadLookups();
		}
		if (!open) {
			lookupLoadStarted = false;
		}
	});

	async function loadLookups() {
		lookupLoading = true;
		lookupError = '';
		try {
			const [nextDepartments, nextLocations] = await Promise.all([getDepartments(), getLocations()]);
			departments = nextDepartments;
			locations = nextLocations;
		} catch (error) {
			if (error instanceof LookupApiError && error.kind === 'unauthenticated') {
				await goto('/login', { replaceState: true });
				return;
			}
			lookupError = 'Unable to load departments and locations. Please try again.';
		} finally {
			lookupLoading = false;
		}
	}

	function handleCancel() {
		if (submitting) {
			return;
		}
		resetForm();
		onClose();
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (disabled || submitting || lookupLoading || departments.length === 0) {
			return;
		}

		const request = buildRequest();
		if (!request) {
			return;
		}

		submitting = true;
		submitError = '';
		try {
			await createTicket(request);
			resetForm();
			onClose();
			await onCreated();
		} catch (error) {
			if (error instanceof TicketApiError && error.kind === 'unauthenticated') {
				await goto('/login', { replaceState: true });
				return;
			}
			if (error instanceof TicketApiError && error.kind === 'invalid_input') {
				submitError = createTicketErrorMessage(error.code);
				return;
			}
			submitError = 'The request could not be created. Please try again.';
		} finally {
			submitting = false;
		}
	}

	function buildRequest(): CreateTicketRequest | null {
		validationError = '';
		submitError = '';
		const selectedDepartmentID = Number(departmentID);
		if (!Number.isSafeInteger(selectedDepartmentID) || selectedDepartmentID <= 0) {
			validationError = 'Select a department.';
			return null;
		}

		const trimmedTitle = title.trim();
		if (!trimmedTitle) {
			validationError = 'Enter a request title.';
			return null;
		}
		if ([...trimmedTitle].length > 255) {
			validationError = 'Request title must be 255 characters or fewer.';
			return null;
		}

		const trimmedDescription = description.trim();
		if ([...trimmedDescription].length > 5000) {
			validationError = 'Description must be 5000 characters or fewer.';
			return null;
		}

		const normalizedDueAt = toRFC3339(dueAt);
		if (dueAt && !normalizedDueAt) {
			validationError = 'Enter a valid due time.';
			return null;
		}

		let selectedLocationID: number | null = null;
		if (locationID) {
			selectedLocationID = Number(locationID);
			if (!Number.isSafeInteger(selectedLocationID) || selectedLocationID <= 0) {
				validationError = 'Select a valid location.';
				return null;
			}
		}

		return {
			department_id: selectedDepartmentID,
			location_id: selectedLocationID,
			title: trimmedTitle,
			description: trimmedDescription || null,
			priority,
			due_at: normalizedDueAt
		};
	}

	function toRFC3339(value: string): string | null {
		if (!value) {
			return null;
		}
		const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})$/.exec(value);
		if (!match) {
			return null;
		}
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) {
			return null;
		}
		if (
			date.getFullYear() !== Number(match[1]) ||
			date.getMonth() !== Number(match[2]) - 1 ||
			date.getDate() !== Number(match[3]) ||
			date.getHours() !== Number(match[4]) ||
			date.getMinutes() !== Number(match[5])
		) {
			return null;
		}
		return date.toISOString();
	}

	function resetForm() {
		departmentID = '';
		locationID = '';
		title = '';
		description = '';
		priority = false;
		dueAt = '';
		validationError = '';
		submitError = '';
	}

	function createTicketErrorMessage(code?: CreateTicketErrorCode): string {
		if (code === 'department_unavailable') {
			return 'The selected department is no longer available. Please select another department.';
		}
		if (code === 'location_unavailable') {
			return 'The selected location is no longer available. Please select another location.';
		}
		return 'Please review the request details and try again.';
	}
</script>

{#if open}
	<button class="new-request-backdrop" type="button" aria-label="Close New Request" onclick={handleCancel}></button>
	<div class="new-request-dialog" role="dialog" aria-modal="true" aria-labelledby="new-request-title">
		<section class="new-request-panel">
			<header class="new-request-header">
				<div>
					<p class="eyebrow">Tickets</p>
					<h2 id="new-request-title">New Request</h2>
				</div>
				<button class="new-request-close" type="button" aria-label="Close New Request" disabled={submitting} onclick={handleCancel}>×</button>
			</header>

			{#if lookupLoading}
				<p class="new-request-status" role="status" aria-live="polite">Loading departments and locations…</p>
			{:else if lookupError}
				<div class="new-request-error" role="alert" aria-live="assertive">
					<span>{lookupError}</span>
					<button class="secondary-button" type="button" onclick={() => void loadLookups()}>Retry</button>
				</div>
			{:else}
				<form class="new-request-form" onsubmit={handleSubmit} aria-busy={submitting}>
					<div class="new-request-field">
						<label for="new-request-department">Department <span aria-hidden="true">*</span></label>
						<select id="new-request-department" name="department_id" bind:value={departmentID} required disabled={submitting || departments.length === 0}>
							<option value="">Select a department</option>
							{#each departments as department (department.id)}
								<option value={String(department.id)}>{department.name}</option>
							{/each}
						</select>
						{#if departments.length === 0}<span class="new-request-help">No active departments are available.</span>{/if}
					</div>

					<div class="new-request-field">
						<label for="new-request-location">Location</label>
						<select id="new-request-location" name="location_id" bind:value={locationID} disabled={submitting}>
							<option value="">No location</option>
							{#each locations as location (location.id)}
								<option value={String(location.id)}>{location.name}</option>
							{/each}
						</select>
					</div>

					<div class="new-request-field">
						<label for="new-request-title-field">Request <span aria-hidden="true">*</span></label>
						<input id="new-request-title-field" name="title" type="text" maxlength="255" bind:value={title} required disabled={submitting} />
					</div>

					<div class="new-request-field">
						<label for="new-request-description">Description</label>
						<textarea id="new-request-description" name="description" maxlength="5000" rows="5" bind:value={description} disabled={submitting}></textarea>
					</div>

					<div class="new-request-options">
						<label class="new-request-checkbox">
							<input type="checkbox" name="priority" bind:checked={priority} disabled={submitting} />
							<span>Priority</span>
						</label>
						<div class="new-request-field new-request-due-field">
							<label for="new-request-due">Due time</label>
							<input id="new-request-due" name="due_at" type="datetime-local" bind:value={dueAt} disabled={submitting} />
						</div>
					</div>

					{#if validationError}<p class="new-request-error" role="alert" aria-live="assertive">{validationError}</p>{/if}
					{#if submitError}<p class="new-request-error" role="alert" aria-live="assertive">{submitError}</p>{/if}

					<div class="new-request-actions">
						<button class="secondary-button" type="button" disabled={submitting} onclick={handleCancel}>Cancel</button>
						<button class="primary-button" type="submit" disabled={disabled || submitting || departments.length === 0}>
							{submitting ? 'Creating…' : 'Create request'}
						</button>
					</div>
				</form>
			{/if}
		</section>
	</div>
{/if}
