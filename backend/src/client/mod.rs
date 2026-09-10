//! Client-facing feature modules.
//!
//! Client workflows are grouped under this namespace so the application can
//! be built and evolved for staff users first without mixing them with
//! administrator-only controls or shared infrastructure.

pub mod announcements;
pub mod auth;
pub mod chat;
pub mod checklist;
pub mod reports;
pub mod settings;
pub mod staff_meal;
pub mod tickets;
