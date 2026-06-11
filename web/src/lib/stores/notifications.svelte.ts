// ─── Types ───────────────────────────────────────────────────────────────────

export type NotificationType = 'success' | 'error' | 'warning' | 'info';
export type NotificationPosition = 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left';

export interface Notification {
	id: string;
	type: NotificationType;
	message: string;
	duration?: number; // ms, default 4000, 0 = persistent
	dismissible?: boolean; // default true
}

export interface NotificationOptions {
	duration?: number;
	dismissible?: boolean;
}

// ─── State ───────────────────────────────────────────────────────────────────

let list = $state<Notification[]>([]);

// ─── Mutations ───────────────────────────────────────────────────────────────

function addNotification(
	type: NotificationType,
	message: string,
	options?: NotificationOptions
): string {
	const id = crypto.randomUUID();
	const notification: Notification = {
		id,
		type,
		message,
		duration: options?.duration ?? 4000,
		dismissible: options?.dismissible ?? true
	};
	list.push(notification);
	return id;
}

export function success(message: string, options?: NotificationOptions): string {
	return addNotification('success', message, options);
}

export function error(message: string, options?: NotificationOptions): string {
	return addNotification('error', message, options);
}

export function warning(message: string, options?: NotificationOptions): string {
	return addNotification('warning', message, options);
}

export function info(message: string, options?: NotificationOptions): string {
	return addNotification('info', message, options);
}

export function dismiss(id: string): void {
	list = list.filter((n) => n.id !== id);
}

export function clear(): void {
	list = [];
}

// ─── Getters ─────────────────────────────────────────────────────────────────

export function getNotifications(): Notification[] {
	return list;
}
