-- get_subscription returns the details of the provided subscription as a json
-- object.
create or replace function get_subscription(p_subscription jsonb)
returns setof json as $$
    select json_build_object(
        'kind', p.package_kind_id,
        'name', p.name,
        'normalized_name', p.normalized_name,
        'logo_image_id', p.logo_image_id,
        'user_alias', u.alias,
        'organization_name', o.name,
        'organization_display_name', o.display_name,
        'chart_repository', (select nullif(
            jsonb_build_object(
                'name', r.name,
                'display_name', r.display_name
            ),
            '{"name": null, "display_name": null}'::jsonb
        )),
        'notification_kind', s.notification_kind_id
    )
    from package p
    join subscription s using (package_id)
    left join chart_repository r using (chart_repository_id)
    left join "user" u on p.user_id = u.user_id or r.user_id = u.user_id
    left join organization o
        on p.organization_id = o.organization_id or r.organization_id = o.organization_id
    where s.user_id = (p_subscription->>'user_id')::uuid
    and s.package_id = (p_subscription->>'package_id')::uuid
    and s.notification_kind_id = (p_subscription->>'notification_kind')::int;
$$ language sql;
