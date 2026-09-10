use std::{env, str::FromStr, time::Duration};

use thiserror::Error;
use url::Url;

const DEFAULT_MAX_CONNECTIONS: u32 = 10;
const DEFAULT_MIN_CONNECTIONS: u32 = 1;
const DEFAULT_ACQUIRE_TIMEOUT_SECONDS: u64 = 5;

#[derive(Debug, Error)]
pub enum ConfigError {
    #[error("missing environment variable {name}")]
    MissingVariable { name: &'static str },
    #[error("invalid value for environment variable {name}")]
    InvalidValue { name: &'static str },
    #[error("database URL is invalid")]
    InvalidDatabaseUrl,
    #[error("database URL component is invalid")]
    InvalidUrlComponent,
    #[error("database pool minimum cannot exceed maximum")]
    InvalidPoolBounds,
    #[error("development admin seed configuration is invalid")]
    InvalidSeedAdmin,
}

pub struct Config {
    database: DatabaseSettings,
    seed_admin: Option<DevelopmentAdmin>,
}

impl Config {
    pub fn from_env() -> Result<Self, ConfigError> {
        let max_connections =
            parse_env_or_default("DATABASE_MAX_CONNECTIONS", DEFAULT_MAX_CONNECTIONS)?;
        let min_connections =
            parse_env_or_default("DATABASE_MIN_CONNECTIONS", DEFAULT_MIN_CONNECTIONS)?;
        let acquire_timeout_seconds = parse_env_or_default(
            "DATABASE_ACQUIRE_TIMEOUT_SECONDS",
            DEFAULT_ACQUIRE_TIMEOUT_SECONDS,
        )?;

        let database_url = non_empty_env("DATABASE_URL");
        let pool = PoolSettings::new(max_connections, min_connections, acquire_timeout_seconds);
        let database = match database_url {
            Some(url) => DatabaseSettings::from_url(url, pool)?,
            None => {
                let host = required_env("DATABASE_HOST")?;
                let port = parse_required_env("DATABASE_PORT")?;
                let name = required_env("DATABASE_NAME")?;
                let user = required_env("DATABASE_USER")?;
                let password = required_env("DATABASE_PASSWORD")?;

                DatabaseSettings::from_parts_with_pool_values(
                    &host, port, &name, &user, &password, pool,
                )?
            }
        };

        let app_env = env::var("APP_ENV").unwrap_or_else(|_| "development".to_owned());
        let seed_enabled = parse_bool_or_default("SEED_DEVELOPMENT_DATA", false)?;
        let seed_admin = if app_env.eq_ignore_ascii_case("development") && seed_enabled {
            Some(DevelopmentAdmin::from_env()?)
        } else {
            None
        };

        Ok(Self {
            database,
            seed_admin,
        })
    }

    pub fn database(&self) -> &DatabaseSettings {
        &self.database
    }

    pub fn seed_admin(&self) -> Option<&DevelopmentAdmin> {
        self.seed_admin.as_ref()
    }
}

pub struct DatabaseSettings {
    url: String,
    max_connections: u32,
    min_connections: u32,
    acquire_timeout: Duration,
}

#[derive(Clone, Copy)]
struct PoolSettings {
    max_connections: u32,
    min_connections: u32,
    acquire_timeout_seconds: u64,
}

impl PoolSettings {
    fn new(max_connections: u32, min_connections: u32, acquire_timeout_seconds: u64) -> Self {
        Self {
            max_connections,
            min_connections,
            acquire_timeout_seconds,
        }
    }
}

impl DatabaseSettings {
    pub fn from_parts(
        host: &str,
        port: u16,
        name: &str,
        user: &str,
        password: &str,
    ) -> Result<Self, ConfigError> {
        Self::from_parts_with_pool_values(
            host,
            port,
            name,
            user,
            password,
            PoolSettings::new(
                DEFAULT_MAX_CONNECTIONS,
                DEFAULT_MIN_CONNECTIONS,
                DEFAULT_ACQUIRE_TIMEOUT_SECONDS,
            ),
        )
    }

    fn from_parts_with_pool_values(
        host: &str,
        port: u16,
        name: &str,
        user: &str,
        password: &str,
        pool: PoolSettings,
    ) -> Result<Self, ConfigError> {
        if host.is_empty() || name.is_empty() || user.is_empty() {
            return Err(ConfigError::InvalidValue {
                name: "DATABASE_HOST/DATABASE_NAME/DATABASE_USER",
            });
        }

        let mut url =
            Url::parse("postgres://localhost").map_err(|_| ConfigError::InvalidDatabaseUrl)?;
        url.set_host(Some(host))
            .map_err(|_| ConfigError::InvalidUrlComponent)?;
        url.set_port(Some(port))
            .map_err(|_| ConfigError::InvalidUrlComponent)?;
        url.set_username(user)
            .map_err(|_| ConfigError::InvalidUrlComponent)?;
        url.set_password(Some(password))
            .map_err(|_| ConfigError::InvalidUrlComponent)?;
        url.set_path(&format!("/{name}"));

        Self::from_url(url.to_string(), pool)
    }

    pub fn with_pool_values(
        url: impl Into<String>,
        max_connections: u32,
        min_connections: u32,
        acquire_timeout_seconds: u64,
    ) -> Result<Self, ConfigError> {
        Self::from_url(
            url.into(),
            PoolSettings::new(max_connections, min_connections, acquire_timeout_seconds),
        )
    }

    fn from_url(url: String, pool: PoolSettings) -> Result<Self, ConfigError> {
        Url::parse(&url).map_err(|_| ConfigError::InvalidDatabaseUrl)?;

        if pool.max_connections == 0 || pool.min_connections > pool.max_connections {
            return Err(ConfigError::InvalidPoolBounds);
        }

        if pool.acquire_timeout_seconds == 0 {
            return Err(ConfigError::InvalidValue {
                name: "DATABASE_ACQUIRE_TIMEOUT_SECONDS",
            });
        }

        Ok(Self {
            url,
            max_connections: pool.max_connections,
            min_connections: pool.min_connections,
            acquire_timeout: Duration::from_secs(pool.acquire_timeout_seconds),
        })
    }

    pub fn url(&self) -> &str {
        &self.url
    }

    pub fn max_connections(&self) -> u32 {
        self.max_connections
    }

    pub fn min_connections(&self) -> u32 {
        self.min_connections
    }

    pub fn acquire_timeout(&self) -> Duration {
        self.acquire_timeout
    }
}

pub struct DevelopmentAdmin {
    username: String,
    employee_code: String,
    full_name: String,
    department_code: String,
    password: String,
}

impl DevelopmentAdmin {
    fn from_env() -> Result<Self, ConfigError> {
        let password = required_env("SEED_ADMIN_PASSWORD")?;
        let username =
            non_empty_env("SEED_ADMIN_USERNAME").unwrap_or_else(|| "hothienty".to_owned());
        let employee_code =
            non_empty_env("SEED_ADMIN_EMPLOYEE_CODE").unwrap_or_else(|| "IT-ADMIN-001".to_owned());
        let full_name =
            non_empty_env("SEED_ADMIN_FULL_NAME").unwrap_or_else(|| "Ho Thien Ty".to_owned());
        let department_code =
            non_empty_env("SEED_ADMIN_DEPARTMENT_CODE").unwrap_or_else(|| "IT".to_owned());

        if username.is_empty()
            || employee_code.is_empty()
            || full_name.is_empty()
            || department_code.is_empty()
            || password.is_empty()
        {
            return Err(ConfigError::InvalidSeedAdmin);
        }

        Ok(Self {
            username,
            employee_code,
            full_name,
            department_code,
            password,
        })
    }

    pub fn username(&self) -> &str {
        &self.username
    }

    pub fn employee_code(&self) -> &str {
        &self.employee_code
    }

    pub fn full_name(&self) -> &str {
        &self.full_name
    }

    pub fn department_code(&self) -> &str {
        &self.department_code
    }

    pub fn password(&self) -> &str {
        &self.password
    }
}

fn required_env(name: &'static str) -> Result<String, ConfigError> {
    let value = env::var(name).map_err(|_| ConfigError::MissingVariable { name })?;
    if value.trim().is_empty() {
        return Err(ConfigError::MissingVariable { name });
    }
    Ok(value)
}

fn non_empty_env(name: &str) -> Option<String> {
    env::var(name).ok().filter(|value| !value.trim().is_empty())
}

fn parse_env_or_default<T>(name: &'static str, default: T) -> Result<T, ConfigError>
where
    T: FromStr,
{
    match non_empty_env(name) {
        Some(value) => value
            .parse()
            .map_err(|_| ConfigError::InvalidValue { name }),
        None => Ok(default),
    }
}

fn parse_required_env<T>(name: &'static str) -> Result<T, ConfigError>
where
    T: FromStr,
{
    let value = required_env(name)?;
    value
        .parse()
        .map_err(|_| ConfigError::InvalidValue { name })
}

fn parse_bool_or_default(name: &'static str, default: bool) -> Result<bool, ConfigError> {
    match non_empty_env(name) {
        Some(value) => match value.to_ascii_lowercase().as_str() {
            "1" | "true" | "yes" | "on" => Ok(true),
            "0" | "false" | "no" | "off" => Ok(false),
            _ => Err(ConfigError::InvalidValue { name }),
        },
        None => Ok(default),
    }
}

#[cfg(test)]
mod tests {
    use super::DatabaseSettings;

    #[test]
    fn split_database_settings_encode_credentials_and_hyphenated_database_names() {
        let settings =
            DatabaseSettings::from_parts("127.0.0.1", 5432, "bwp-sonasea", "postgres", "p@ss:word")
                .expect("valid database settings");

        assert_eq!(
            settings.url(),
            "postgres://postgres:p%40ss%3Aword@127.0.0.1:5432/bwp-sonasea"
        );
    }

    #[test]
    fn database_settings_reject_minimum_pool_size_above_maximum() {
        let result = DatabaseSettings::with_pool_values("postgres://localhost/db", 8, 9, 5);

        assert!(result.is_err());
    }
}
