//! Authentication boundary for username-only administrative access.
//!
//! Accounts are provisioned by an administrator. Email login, self-service
//! registration, and password-reset flows are intentionally out of scope.

pub mod handler;
pub mod model;
pub mod repository;
pub mod service;
