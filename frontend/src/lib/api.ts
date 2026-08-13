// Tauri backend bindings.
//
// This module replaces the auto-generated Wails bindings (`wailsjs/`). Each
// function wraps a Tauri command defined in `src-tauri/src/lib.rs`. The command
// implementations are currently stubs — the website blocking / daemon logic is
// not wired up yet.

import { invoke } from '@tauri-apps/api/core';
import { openUrl } from '@tauri-apps/plugin-opener';

export interface AppSettings {
	unblockWaiting: number;
}

export interface EnvironmentInfo {
	platform: string;
	arch: string;
}

export function ConnectToDaemon(): Promise<string> {
	return invoke('connect_to_daemon');
}

export function CheckBlocking(): Promise<boolean> {
	return invoke('check_blocking');
}

export function CheckDaemonInstalled(): Promise<boolean> {
	return invoke('check_daemon_installed');
}

export function InstallAndStartDaemon(): Promise<string> {
	return invoke('install_and_start_daemon');
}

export function SendBlockList(list: string): Promise<boolean> {
	return invoke('send_block_list', { list });
}

export function StartBlocking(): Promise<boolean> {
	return invoke('start_blocking');
}

export function StopBlocking(): Promise<string> {
	return invoke('stop_blocking');
}

export function LoadBlockedWebsites(): Promise<string> {
	return invoke('load_blocked_websites');
}

export function SaveBlockedWebsites(json: string): Promise<boolean> {
	return invoke('save_blocked_websites', { json });
}

export function LoadSettings(): Promise<AppSettings> {
	return invoke('load_settings');
}

export function SaveSettings(settings: AppSettings): Promise<boolean> {
	return invoke('save_settings', { settings });
}

export function Environment(): Promise<EnvironmentInfo> {
	return invoke('environment');
}

export function BrowserOpenURL(url: string): Promise<void> {
	return openUrl(url);
}
