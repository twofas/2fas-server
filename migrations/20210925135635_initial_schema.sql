-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `browser_extensions` (
    `id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    `browser_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    `browser_version` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    `public_key` varchar(768) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    `created_at` timestamp NOT NULL,
    `updated_at` timestamp NULL DEFAULT NULL,
    `deleted_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `browser_extensions_2fa_requests` (
    `id` char(36) NOT NULL,
    `extension_id` char(36) NOT NULL,
    `domain` varchar(256) NOT NULL,
    `status` enum('pending','completed','terminated') NOT NULL,
    `created_at` timestamp NOT NULL,
    `updated_at` timestamp NULL DEFAULT NULL,
    `deleted_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `icons_requests` (
    `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
    `caller_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL,
    `service_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
    `issuers` json NOT NULL,

    `description` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL,
    `light_icon_url` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL,
    `dark_icon_url` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL,
    `created_at` timestamp NOT NULL,
    `updated_at` timestamp NULL DEFAULT NULL,
    `deleted_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `icons` (
    `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
    `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
    `url` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL,
    `width` smallint NOT NULL,
    `height` smallint NOT NULL,
    `type` enum('light','dark') COLLATE utf8mb4_unicode_ci NOT NULL,
    `created_at` timestamp NOT NULL,
    `updated_at` timestamp NULL DEFAULT NULL,
    `deleted_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `icons_collections` (
    `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
    `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
    `description` tinytext COLLATE utf8mb4_unicode_ci,
    `icons` json DEFAULT NULL,
    `created_at` timestamp NOT NULL,
    `updated_at` timestamp NULL DEFAULT NULL,
    `deleted_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `mobile_debug_logs_audit` (
    `id` char(36) NOT NULL,
    `username` varchar(128) NOT NULL,
    `file` varchar(255) DEFAULT NULL,
    `description` text,
    `created_at` timestamp NOT NULL,
    `updated_at` timestamp NULL DEFAULT NULL,
    `deleted_at` timestamp NULL DEFAULT NULL,
    `expire_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `mobile_devices` (
    `id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    `name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    `platform` enum('android','ios','huawei') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `fcm_token` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    `created_at` timestamp NOT NULL,
    `updated_at` timestamp NULL DEFAULT NULL,
    `deleted_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `mobile_device_browser_extension` (
    `device_id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    `extension_id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    `created_at` timestamp NOT NULL,
    `updated_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`device_id`,`extension_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `mobile_notifications` (
    `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
    `icon` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
    `link` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
    `message` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL,
    `push` tinyint(1) NOT NULL,
    `platform` enum('android','ios','huawei') COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `version` varchar(12) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `created_at` timestamp NOT NULL,
    `updated_at` timestamp NULL DEFAULT NULL,
    `deleted_at` timestamp NULL DEFAULT NULL,
    `published_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `web_services` (
    `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
    `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
    `description` mediumtext COLLATE utf8mb4_unicode_ci,
    `issuers` json NOT NULL,
    `tags` json NOT NULL,
    `icons_collections` json NOT NULL,
    `match_rules` json DEFAULT NULL,
    `created_at` timestamp NOT NULL,
    `updated_at` timestamp NULL DEFAULT NULL,
    `deleted_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE OR REPLACE VIEW web_services_dump AS

SELECT
    JSON_OBJECT(
            "id", ws.id,
            "name", ws.name,
            "issuers", ws.issuers,
            "tags", ws.tags,
            "match_rules", ws.match_rules,
            "icons_collections", IF(JSON_LENGTH(icons_collections) > 0, JSON_ARRAYAGG(
            JSON_OBJECT(
                    "id", ic.id,
                    "name", ic.name,
                    "icons", IF(JSON_LENGTH(ic.icons) > 0,
                                (
                                    SELECT JSON_ARRAYAGG(
                                                   JSON_OBJECT(
                                                           "id", i.id,
                                                           "name", i.name,
                                                           "url", i.url,
                                                           "type", i.type,
                                                           "width", i.width,
                                                           "height", i.height,
                                                           "created_at", i.created_at,
                                                           "updated_at", i.updated_at
                                                       )
                                               ) FROM icons i WHERE i.id MEMBER OF (ic.icons)
                                ), JSON_ARRAY())
                )), JSON_ARRAY()),
            "created_at", ws.created_at,
            "updated_at", ws.updated_at
        ) as web_service

FROM web_services ws

         LEFT JOIN icons_collections ic ON ic.id MEMBER OF (ws.icons_collections)

WHERE ws.deleted_at IS NULL

GROUP BY ws.id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
