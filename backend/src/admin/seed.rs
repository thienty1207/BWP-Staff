use argon2::{
    Algorithm, Argon2, Params, Version,
    password_hash::{PasswordHasher, SaltString, rand_core::OsRng},
};
use sqlx::{PgPool, query, query_scalar};
use thiserror::Error;

use crate::config::DevelopmentAdmin;

#[derive(Debug, Error)]
pub enum SeedError {
    #[error("password hashing failed")]
    PasswordHash,
    #[error("database seed query failed")]
    Database(#[from] sqlx::Error),
    #[error("required development department is missing")]
    MissingDepartment,
}

#[derive(Debug, PartialEq, Eq)]
pub enum SeedOutcome {
    Created,
    AlreadyPresent,
}

pub fn hash_password(password: &str) -> Result<String, SeedError> {
    let salt = SaltString::generate(&mut OsRng);
    let argon2 = Argon2::new(Algorithm::Argon2id, Version::V0x13, Params::default());

    argon2
        .hash_password(password.as_bytes(), &salt)
        .map(|hash| hash.to_string())
        .map_err(|_| SeedError::PasswordHash)
}

pub async fn seed_development_admin(
    pool: &PgPool,
    admin: &DevelopmentAdmin,
) -> Result<SeedOutcome, SeedError> {
    let mut transaction = pool.begin().await?;
    let existing_user: Option<i64> = query_scalar("SELECT id FROM users WHERE username = $1")
        .bind(admin.username())
        .fetch_optional(&mut *transaction)
        .await?;

    if existing_user.is_some() {
        return Ok(SeedOutcome::AlreadyPresent);
    }

    let department_id: i64 =
        query_scalar("SELECT id FROM departments WHERE code = $1 AND is_active = TRUE")
            .bind(admin.department_code())
            .fetch_optional(&mut *transaction)
            .await?
            .ok_or(SeedError::MissingDepartment)?;

    let password_hash = hash_password(admin.password())?;
    let user_id: Option<i64> = query_scalar(
        r#"
        INSERT INTO users (
            username,
            employee_code,
            email,
            password_hash,
            full_name,
            department_id,
            role
        )
        VALUES ($1, $2, NULL, $3, $4, $5, 'admin'::user_role)
        ON CONFLICT (username) DO NOTHING
        RETURNING id
        "#,
    )
    .bind(admin.username())
    .bind(admin.employee_code())
    .bind(password_hash)
    .bind(admin.full_name())
    .bind(department_id)
    .fetch_optional(&mut *transaction)
    .await?;

    let Some(user_id) = user_id else {
        return Ok(SeedOutcome::AlreadyPresent);
    };

    query(
        r#"
        INSERT INTO user_preferences (user_id)
        VALUES ($1)
        ON CONFLICT (user_id) DO NOTHING
        "#,
    )
    .bind(user_id)
    .execute(&mut *transaction)
    .await?;

    transaction.commit().await?;

    Ok(SeedOutcome::Created)
}

#[cfg(test)]
mod tests {
    use argon2::{Argon2, PasswordHash, PasswordVerifier};

    use super::hash_password;

    #[test]
    fn password_hash_is_argon2id_and_verifies_the_original_password() {
        let hash = hash_password("test-only-password").expect("hash password");
        let parsed = PasswordHash::new(&hash).expect("parse password hash");

        Argon2::default()
            .verify_password(b"test-only-password", &parsed)
            .expect("password verifies");
        assert_eq!(parsed.algorithm.as_str(), "argon2id");
    }
}
