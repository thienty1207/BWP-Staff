use std::time::{SystemTime, UNIX_EPOCH};

use argon2::{Argon2, PasswordHash, PasswordVerifier};
use backend::{
    config::Config,
    seed::{SeedOutcome, seed_development_admin},
    shared::database,
};
use sqlx::{PgPool, Row};

fn unique_test_suffix() -> String {
    let nanos = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("system clock is after unix epoch")
        .as_nanos();
    format!("{}-{nanos}", std::process::id())
}

async fn database_pool() -> (PgPool, Config) {
    dotenvy::dotenv().ok();
    let config = Config::from_env().expect("local PostgreSQL configuration is valid");
    let pool = database::connect(config.database())
        .await
        .expect("local PostgreSQL is reachable");
    database::run_migrations(&pool)
        .await
        .expect("SPEC-01 migrations run successfully");
    (pool, config)
}

async fn enum_values(pool: &PgPool, type_name: &str) -> Vec<String> {
    sqlx::query_scalar(
        r#"
        SELECT enumlabel::text
        FROM pg_enum
        JOIN pg_type ON pg_type.oid = pg_enum.enumtypid
        WHERE pg_type.typname = $1
        ORDER BY enumsortorder
        "#,
    )
    .bind(type_name)
    .fetch_all(pool)
    .await
    .expect("query enum values")
}

#[tokio::test]
async fn spec01_database_contract_is_valid() {
    let (pool, config) = database_pool().await;
    let required_tables = vec![
        "departments".to_owned(),
        "users".to_owned(),
        "user_preferences".to_owned(),
        "auth_sessions".to_owned(),
        "locations".to_owned(),
        "tickets".to_owned(),
        "ticket_activity".to_owned(),
        "ticket_messages".to_owned(),
        "message_attachments".to_owned(),
        "checklist_items".to_owned(),
        "announcements".to_owned(),
        "staff_meals".to_owned(),
        "notifications".to_owned(),
        "audit_logs".to_owned(),
    ];

    let table_count: i64 = sqlx::query_scalar(
        r#"
        SELECT COUNT(*)::bigint
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = ANY($1)
        "#,
    )
    .bind(&required_tables)
    .fetch_one(&pool)
    .await
    .expect("query required table count");
    assert_eq!(table_count as usize, required_tables.len());

    assert_eq!(enum_values(&pool, "user_role").await, ["admin", "staff"]);
    assert_eq!(
        enum_values(&pool, "ticket_status").await,
        ["pending", "accepted", "closed"]
    );
    assert_eq!(
        enum_values(&pool, "notification_type").await,
        [
            "ticket_created",
            "ticket_accepted",
            "ticket_assigned",
            "ticket_closed",
            "new_message",
            "announcement",
            "system",
        ]
    );
    assert_eq!(
        enum_values(&pool, "audit_action").await,
        [
            "create",
            "update",
            "delete",
            "login",
            "logout",
            "accept",
            "assign",
            "close",
            "publish",
            "unpublish",
        ]
    );

    let username_not_null: bool = sqlx::query_scalar(
        r#"
        SELECT is_nullable = 'NO'
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'username'
        "#,
    )
    .fetch_one(&pool)
    .await
    .expect("username column exists");
    assert!(username_not_null);

    let email_nullable: bool = sqlx::query_scalar(
        r#"
        SELECT is_nullable = 'YES'
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'email'
        "#,
    )
    .fetch_one(&pool)
    .await
    .expect("email column exists");
    assert!(email_nullable);

    let required_indexes = vec![
        "idx_tickets_status_created_at".to_owned(),
        "idx_tickets_requester_created_at".to_owned(),
        "idx_tickets_assigned_status".to_owned(),
        "idx_tickets_department_status_created".to_owned(),
        "idx_tickets_location_created_at".to_owned(),
        "idx_ticket_messages_ticket_created".to_owned(),
        "idx_notifications_user_unread_created".to_owned(),
        "idx_announcements_published_at".to_owned(),
        "idx_announcements_author_created_at".to_owned(),
        "idx_auth_sessions_active_user_expires_at".to_owned(),
        "idx_staff_meals_valid_range".to_owned(),
        "idx_audit_logs_entity_created".to_owned(),
    ];
    let index_count: i64 = sqlx::query_scalar(
        r#"
        SELECT COUNT(*)::bigint
        FROM pg_indexes
        WHERE schemaname = 'public' AND indexname = ANY($1)
        "#,
    )
    .bind(&required_indexes)
    .fetch_one(&pool)
    .await
    .expect("query required indexes");
    assert_eq!(index_count as usize, required_indexes.len());

    let department_codes: Vec<String> =
        sqlx::query_scalar("SELECT code FROM departments WHERE code = ANY($1) ORDER BY code")
            .bind(vec!["ENG", "FB", "FO", "HK", "HR", "IT"])
            .fetch_all(&pool)
            .await
            .expect("query development departments");
    assert_eq!(department_codes, ["ENG", "FB", "FO", "HK", "HR", "IT"]);

    let location_count: i64 = sqlx::query_scalar(
        r#"
        SELECT COUNT(*)::bigint
        FROM locations
        WHERE name = ANY($1)
        "#,
    )
    .bind(vec![
        "Lobby",
        "Ballroom",
        "Back Office",
        "Room 8020",
        "Room 7309",
        "Villa",
    ])
    .fetch_one(&pool)
    .await
    .expect("query development locations");
    assert_eq!(location_count, 6);

    let admin = config
        .seed_admin()
        .expect("development admin seed is enabled in local env");
    let first_seed = seed_development_admin(&pool, admin)
        .await
        .expect("development admin seed succeeds");
    assert!(matches!(
        first_seed,
        SeedOutcome::Created | SeedOutcome::AlreadyPresent
    ));

    let before_hash: String =
        sqlx::query_scalar("SELECT password_hash FROM users WHERE username = 'hothienty'")
            .fetch_one(&pool)
            .await
            .expect("development admin exists");
    assert!(before_hash.starts_with("$argon2id$"));

    let parsed_hash = PasswordHash::new(&before_hash).expect("development hash is valid PHC");
    Argon2::default()
        .verify_password(admin.password().as_bytes(), &parsed_hash)
        .expect("development admin password verifies");

    let admin_row = sqlx::query(
        r#"
        SELECT u.role::text AS role, u.email, d.code AS department_code
        FROM users u
        JOIN departments d ON d.id = u.department_id
        WHERE u.username = 'hothienty'
        "#,
    )
    .fetch_one(&pool)
    .await
    .expect("development admin row is queryable");
    assert_eq!(admin_row.get::<String, _>("role"), "admin");
    assert!(admin_row.get::<Option<String>, _>("email").is_none());
    assert_eq!(admin_row.get::<String, _>("department_code"), "IT");

    let second_seed = seed_development_admin(&pool, admin)
        .await
        .expect("second development admin seed succeeds");
    assert_eq!(second_seed, SeedOutcome::AlreadyPresent);
    let after_hash: String =
        sqlx::query_scalar("SELECT password_hash FROM users WHERE username = 'hothienty'")
            .fetch_one(&pool)
            .await
            .expect("development admin remains queryable");
    assert_eq!(before_hash, after_hash);

    let updated_by_trigger: bool = sqlx::query_scalar(
        r#"
        UPDATE departments
        SET updated_at = created_at
        WHERE code = 'IT'
        RETURNING updated_at > created_at
        "#,
    )
    .fetch_one(&pool)
    .await
    .expect("updated_at trigger runs");
    assert!(updated_by_trigger);

    let suffix = unique_test_suffix();
    let department_id: i64 = sqlx::query_scalar("SELECT id FROM departments WHERE code = 'IT'")
        .fetch_one(&pool)
        .await
        .expect("IT department exists");

    let mut transaction = pool.begin().await.expect("begin FK test transaction");
    let invalid_fk = sqlx::query(
        r#"
        INSERT INTO users (username, employee_code, password_hash, full_name, department_id)
        VALUES ($1, $2, '$argon2id$test', 'Invalid FK Test', 9223372036854775807)
        "#,
    )
    .bind(format!("invalid-fk-{suffix}"))
    .bind(format!("INVALID-FK-{suffix}"))
    .execute(&mut *transaction)
    .await;
    assert!(invalid_fk.is_err());
    transaction.rollback().await.expect("rollback FK test");

    let mut transaction = pool
        .begin()
        .await
        .expect("begin uniqueness test transaction");
    sqlx::query(
        r#"
        INSERT INTO users (username, employee_code, email, password_hash, full_name, department_id)
        VALUES ($1, $2, $3, '$argon2id$test', 'Unique Test', $4)
        "#,
    )
    .bind(format!("unique-{suffix}"))
    .bind(format!("UNIQUE-{suffix}"))
    .bind(format!("unique-{suffix}@example.invalid"))
    .bind(department_id)
    .execute(&mut *transaction)
    .await
    .expect("insert first unique test user");
    let duplicate_username = sqlx::query(
        r#"
        INSERT INTO users (username, employee_code, email, password_hash, full_name, department_id)
        VALUES ($1, $2, $3, '$argon2id$test', 'Duplicate Test', $4)
        "#,
    )
    .bind(format!("unique-{suffix}"))
    .bind(format!("UNIQUE-OTHER-{suffix}"))
    .bind(format!("other-{suffix}@example.invalid"))
    .bind(department_id)
    .execute(&mut *transaction)
    .await;
    assert!(duplicate_username.is_err());
    transaction
        .rollback()
        .await
        .expect("rollback uniqueness test");

    let mut transaction = pool
        .begin()
        .await
        .expect("begin duplicate email test transaction");
    sqlx::query(
        r#"
        INSERT INTO users (username, employee_code, email, password_hash, full_name, department_id)
        VALUES ($1, $2, $3, '$argon2id$test', 'Email Test', $4)
        "#,
    )
    .bind(format!("email-{suffix}"))
    .bind(format!("EMAIL-{suffix}"))
    .bind(format!("shared-{suffix}@example.invalid"))
    .bind(department_id)
    .execute(&mut *transaction)
    .await
    .expect("insert first email test user");
    let duplicate_email = sqlx::query(
        r#"
        INSERT INTO users (username, employee_code, email, password_hash, full_name, department_id)
        VALUES ($1, $2, $3, '$argon2id$test', 'Duplicate Email Test', $4)
        "#,
    )
    .bind(format!("email-other-{suffix}"))
    .bind(format!("EMAIL-OTHER-{suffix}"))
    .bind(format!("shared-{suffix}@example.invalid"))
    .bind(department_id)
    .execute(&mut *transaction)
    .await;
    assert!(duplicate_email.is_err());
    transaction
        .rollback()
        .await
        .expect("rollback duplicate email test");

    let mut transaction = pool
        .begin()
        .await
        .expect("begin duplicate employee code test transaction");
    sqlx::query(
        r#"
        INSERT INTO users (username, employee_code, email, password_hash, full_name, department_id)
        VALUES ($1, $2, $3, '$argon2id$test', 'Employee Code Test', $4)
        "#,
    )
    .bind(format!("employee-{suffix}"))
    .bind(format!("EMPLOYEE-{suffix}"))
    .bind(format!("employee-{suffix}@example.invalid"))
    .bind(department_id)
    .execute(&mut *transaction)
    .await
    .expect("insert first employee code test user");
    let duplicate_employee_code = sqlx::query(
        r#"
        INSERT INTO users (username, employee_code, email, password_hash, full_name, department_id)
        VALUES ($1, $2, $3, '$argon2id$test', 'Duplicate Employee Code Test', $4)
        "#,
    )
    .bind(format!("employee-other-{suffix}"))
    .bind(format!("EMPLOYEE-{suffix}"))
    .bind(format!("employee-other-{suffix}@example.invalid"))
    .bind(department_id)
    .execute(&mut *transaction)
    .await;
    assert!(duplicate_employee_code.is_err());
    transaction
        .rollback()
        .await
        .expect("rollback duplicate employee code test");

    let mut transaction = pool
        .begin()
        .await
        .expect("begin ticket constraint test transaction");
    let invalid_ticket = sqlx::query(
        r#"
        INSERT INTO tickets (requester_id, department_id, title, status)
        VALUES ((SELECT id FROM users WHERE username = 'hothienty'), $1, 'Invalid status transition', 'accepted')
        "#,
    )
    .bind(department_id)
    .execute(&mut *transaction)
    .await;
    assert!(invalid_ticket.is_err());
    transaction
        .rollback()
        .await
        .expect("rollback ticket constraint test");

    pool.close().await;
}
