-- add_subscription adds the provided subscription to the database.
create or replace function add_subscription(p_subscription jsonb)
returns void as $$
    insert into subscription (
        user_id,
        package_id,
        notification_kind_id
    ) values (
        (p_subscription->>'user_id')::uuid,
        (p_subscription->>'package_id')::uuid,
        (p_subscription->>'notification_kind')::int
    );
$$ language sql;
