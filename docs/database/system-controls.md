# system_controls

`system_controls` is the project-level control table. It stores global switches and management fields such as maintenance mode, minimum client version, force update flags, gameplay feature flags, home notices, and rollout policies.

## Fields

- `control_key`: globally unique key, using dotted names such as `system.maintenance_mode`.
- `control_group`: group name, such as `system`, `client`, `game`, or `notice`.
- `name`: display name for admin pages.
- `description`: explanation for operators.
- `value_type`: one of `bool`, `int`, `decimal`, `string`, or `json`.
- `value_text`: current value for non-JSON settings.
- `value_json`: current value for JSON settings.
- `default_value_text`: default value for non-JSON settings.
- `default_value_json`: default value for JSON settings.
- `enabled`: whether the control is active.
- `is_public`: whether the setting may be returned to clients. Sensitive settings must keep this as `0`.
- `sort_order`: ordering inside admin pages or public config responses.
- `version`: configuration version, incremented when the value changes.
- `effective_from`: optional start time.
- `effective_until`: optional end time.
- `created_by`: admin user ID that created the row.
- `updated_by`: admin user ID that last updated the row.
- `created_at`: creation time.
- `updated_at`: update time.

## Initial Controls

- `system.maintenance_mode`: global maintenance switch.
- `client.min_version`: minimum supported client version.
- `client.force_update`: force client update switch.
- `game.feature_flags`: gameplay feature switches.
- `notice.home_banner`: home notice configuration.
