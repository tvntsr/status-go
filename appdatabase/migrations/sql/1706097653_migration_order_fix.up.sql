CREATE TABLE IF NOT EXISTS "token_preferences" (
    "key" TEXT NOT NULL, "position" INTEGER NOT NULL DEFAULT -1, "group_position" INTEGER NOT NULL DEFAULT -1, "visible" BOOLEAN NOT NULL DEFAULT TRUE, "community_id" TEXT NOT NULL DEFAULT '', "testnet" BOOLEAN NOT NULL DEFAULT FALSE, PRIMARY KEY ("key", "testnet")
);

CREATE TABLE "settings2" (
    address VARCHAR NOT NULL, chaos_mode BOOLEAN DEFAULT false, currency VARCHAR DEFAULT 'usd', current_network VARCHAR NOT NULL, custom_bootnodes BLOB, custom_bootnodes_enabled BLOB, dapps_address VARCHAR NOT NULL, eip1581_address VARCHAR, fleet VARCHAR, hide_home_tooltip BOOLEAN DEFAULT false, installation_id VARCHAR NOT NULL, key_uid VARCHAR NOT NULL, keycard_instance_uid VARCHAR, keycard_paired_on UNSIGNED BIGINT, keycard_pairing VARCHAR, last_updated UNSIGNED BIGINT, log_level VARCHAR, mnemonic VARCHAR, name VARCHAR NOT NULL, networks BLOB NOT NULL, node_config BLOB, notifications_enabled BOOLEAN DEFAULT false, photo_path BLOB NOT NULL, pinned_mailservers BLOB, preferred_name VARCHAR, preview_privacy BOOLEAN DEFAULT false, public_key VARCHAR NOT NULL, remember_syncing_choice BOOLEAN DEFAULT false, signing_phrase VARCHAR NOT NULL, stickers_packs_installed BLOB, stickers_recent_stickers BLOB, syncing_on_mobile_network BOOLEAN DEFAULT false, synthetic_id VARCHAR DEFAULT 'id' PRIMARY KEY, usernames BLOB, wallet_root_address VARCHAR NOT NULL, wallet_set_up_passed BOOLEAN DEFAULT false, wallet_visible_tokens VARCHAR, stickers_packs_pending BLOB, waku_enabled BOOLEAN DEFAULT false, waku_bloom_filter_mode BOOLEAN DEFAULT false, appearance INT NOT NULL DEFAULT 0, remote_push_notifications_enabled BOOLEAN DEFAULT FALSE, send_push_notifications BOOLEAN DEFAULT TRUE, push_notifications_server_enabled BOOLEAN DEFAULT FALSE, push_notifications_from_contacts_only BOOLEAN DEFAULT FALSE, push_notifications_block_mentions BOOLEAN DEFAULT FALSE, webview_allow_permission_requests BOOLEAN DEFAULT FALSE, use_mailservers BOOLEAN DEFAULT TRUE, link_preview_request_enabled BOOLEAN DEFAULT TRUE, link_previews_enabled_sites BLOB, profile_pictures_visibility INT NOT NULL DEFAULT 1, anon_metrics_should_send BOOLEAN DEFAULT false, messages_from_contacts_only BOOLEAN DEFAULT FALSE, default_sync_period INTEGER DEFAULT 777600, -- 9 days
    current_user_status BLOB, send_status_updates BOOLEAN DEFAULT TRUE, gif_recents BLOB, gif_favorites BLOB, opensea_enabled BOOLEAN DEFAULT false, profile_pictures_show_to INT NOT NULL DEFAULT 1, telemetry_server_url VARCHAR NOT NULL DEFAULT "", backup_enabled BOOLEAN DEFAULT TRUE, last_backup INT NOT NULL DEFAULT 0, backup_fetched BOOLEAN DEFAULT FALSE, auto_message_enabled BOOLEAN DEFAULT FALSE, gif_api_key TEXT NOT NULL DEFAULT "", display_name TEXT NOT NULL DEFAULT "", test_networks_enabled BOOLEAN NOT NULL DEFAULT FALSE, mutual_contact_enabled BOOLEAN DEFAULT FALSE, bio TEXT NOT NULL DEFAULT "", mnemonic_removed BOOLEAN NOT NULL DEFAULT FALSE, latest_derived_path INT NOT NULL DEFAULT "0", device_name TEXT NOT NULL DEFAULT "", wallet_accounts_position_change_clock INTEGER NOT NULL DEFAULT 0, profile_migration_needed BOOLEAN NOT NULL DEFAULT FALSE, is_sepolia_enabled BOOLEAN NOT NULL DEFAULT FALSE, url_unfurling_mode INT NOT NULL DEFAULT 1, omit_transfers_history_scan BOOLEAN NOT NULL DEFAULT FALSE, mnemonic_was_not_shown BOOLEAN NOT NULL DEFAULT FALSE, wallet_token_preferences_change_clock INTEGER NOT NULL DEFAULT 0, wallet_token_preferences_group_by_community BOOLEAN NOT NULL DEFAULT FALSE, wallet_show_community_asset_when_sending_tokens INTEGER NOT NULL DEFAULT TRUE, wallet_display_assets_below_balance BOOLEAN NOT NULL DEFAULT FALSE, wallet_display_assets_below_balance_threshold UNSIGNED BIGINT NOT NULL DEFAULT 100000000, wallet_collectible_preferences_change_clock INTEGER NOT NULL DEFAULT 0, wallet_collectible_preferences_group_by_collection BOOLEAN NOT NULL DEFAULT FALSE, wallet_collectible_preferences_group_by_community BOOLEAN NOT NULL DEFAULT FALSE
) WITHOUT ROWID;

INSERT INTO
    settings2 (
        address, chaos_mode, currency, current_network, custom_bootnodes, custom_bootnodes_enabled, dapps_address, eip1581_address, fleet, hide_home_tooltip, installation_id, key_uid, keycard_instance_uid, keycard_paired_on, keycard_pairing, last_updated, log_level, mnemonic, name, networks, node_config, notifications_enabled, photo_path, pinned_mailservers, preferred_name, preview_privacy, public_key, remember_syncing_choice, signing_phrase, stickers_packs_installed, stickers_recent_stickers, syncing_on_mobile_network, synthetic_id, usernames, wallet_root_address, wallet_set_up_passed, wallet_visible_tokens, stickers_packs_pending, waku_enabled, waku_bloom_filter_mode, appearance, remote_push_notifications_enabled, send_push_notifications, push_notifications_server_enabled, push_notifications_from_contacts_only, push_notifications_block_mentions, webview_allow_permission_requests, use_mailservers, link_preview_request_enabled, link_previews_enabled_sites, profile_pictures_visibility, anon_metrics_should_send, messages_from_contacts_only, default_sync_period, current_user_status, send_status_updates, gif_recents, gif_favorites, opensea_enabled, profile_pictures_show_to, telemetry_server_url, backup_enabled, last_backup, backup_fetched, auto_message_enabled, gif_api_key, display_name, test_networks_enabled, mutual_contact_enabled, bio, mnemonic_removed, latest_derived_path, device_name, wallet_accounts_position_change_clock, profile_migration_needed, is_sepolia_enabled, url_unfurling_mode, omit_transfers_history_scan, wallet_show_community_asset_when_sending_tokens, wallet_display_assets_below_balance, wallet_display_assets_below_balance_threshold, wallet_collectible_preferences_change_clock, wallet_collectible_preferences_group_by_collection, wallet_collectible_preferences_group_by_community
    )
SELECT
    address,
    chaos_mode,
    currency,
    current_network,
    custom_bootnodes,
    custom_bootnodes_enabled,
    dapps_address,
    eip1581_address,
    fleet,
    hide_home_tooltip,
    installation_id,
    key_uid,
    keycard_instance_uid,
    keycard_paired_on,
    keycard_pairing,
    last_updated,
    log_level,
    mnemonic,
    name,
    networks,
    node_config,
    notifications_enabled,
    photo_path,
    pinned_mailservers,
    preferred_name,
    preview_privacy,
    public_key,
    remember_syncing_choice,
    signing_phrase,
    stickers_packs_installed,
    stickers_recent_stickers,
    syncing_on_mobile_network,
    synthetic_id,
    usernames,
    wallet_root_address,
    wallet_set_up_passed,
    wallet_visible_tokens,
    stickers_packs_pending,
    waku_enabled,
    waku_bloom_filter_mode,
    appearance,
    remote_push_notifications_enabled,
    send_push_notifications,
    push_notifications_server_enabled,
    push_notifications_from_contacts_only,
    push_notifications_block_mentions,
    webview_allow_permission_requests,
    use_mailservers,
    link_preview_request_enabled,
    link_previews_enabled_sites,
    profile_pictures_visibility,
    anon_metrics_should_send,
    messages_from_contacts_only,
    default_sync_period,
    current_user_status,
    send_status_updates,
    gif_recents,
    gif_favorites,
    opensea_enabled,
    profile_pictures_show_to,
    telemetry_server_url,
    backup_enabled,
    last_backup,
    backup_fetched,
    auto_message_enabled,
    gif_api_key,
    display_name,
    test_networks_enabled,
    mutual_contact_enabled,
    bio,
    mnemonic_removed,
    latest_derived_path,
    device_name,
    wallet_accounts_position_change_clock,
    profile_migration_needed,
    is_sepolia_enabled,
    url_unfurling_mode,
    omit_transfers_history_scan,
    wallet_show_community_asset_when_sending_tokens,
    wallet_display_assets_below_balance,
    wallet_display_assets_below_balance_threshold,
    wallet_collectible_preferences_change_clock,
    wallet_collectible_preferences_group_by_collection,
    wallet_collectible_preferences_group_by_community
FROM settings;

DROP TABLE settings;

ALTER TABLE settings2 RENAME TO settings;

CREATE TABLE "keypairs_accounts2" (
    address VARCHAR PRIMARY KEY, key_uid VARCHAR, pubkey VARCHAR, path VARCHAR NOT NULL DEFAULT "", name VARCHAR NOT NULL DEFAULT "", color VARCHAR NOT NULL DEFAULT "", emoji VARCHAR NOT NULL DEFAULT "", wallet BOOL NOT NULL DEFAULT FALSE, chat BOOL NOT NULL DEFAULT FALSE, hidden BOOL NOT NULL DEFAULT FALSE, operable VARCHAR NOT NULL DEFAULT "no", created_at DATETIME NOT NULL, updated_at DATETIME NOT NULL, clock INT NOT NULL DEFAULT 0, position INT NOT NULL DEFAULT 0, removed BOOLEAN DEFAULT FALSE, prod_preferred_chain_ids VARCHAR NOT NULL DEFAULT "", test_preferred_chain_ids VARCHAR NOT NULL DEFAULT "", address_was_not_shown BOOLEAN NOT NULL DEFAULT FALSE, FOREIGN KEY (key_uid) REFERENCES keypairs (key_uid) ON DELETE CASCADE
);

INSERT INTO
    keypairs_accounts2 (
        address, key_uid, pubkey, path, name, color, emoji, wallet, chat, hidden, operable, created_at, updated_at, clock, position, removed, prod_preferred_chain_ids, test_preferred_chain_ids
    )
SELECT
    address,
    key_uid,
    pubkey,
    path,
    name,
    color,
    emoji,
    wallet,
    chat,
    hidden,
    operable,
    created_at,
    updated_at,
    clock,
    position,
    removed,
    prod_preferred_chain_ids,
    test_preferred_chain_ids
FROM keypairs_accounts;

DROP TABLE keypairs_accounts;

ALTER TABLE keypairs_accounts2 RENAME TO keypairs_accounts;