use anyhow::Result;
use datadog_api_client::datadogV2::api_case_management::{
    CaseManagementAPI, SearchCasesOptionalParams,
};
use datadog_api_client::datadogV2::model::{
    CaseAssign, CaseAssignAttributes, CaseAssignRequest, CaseCreateRequest, CaseEmpty,
    CaseEmptyRequest, CaseNotificationRuleCreateRequest, CaseNotificationRuleUpdateRequest,
    CasePriority, CaseResourceType, CaseStatus, CaseUpdatePriority,
    CaseUpdatePriorityAttributes, CaseUpdatePriorityRequest, CaseUpdateStatus,
    CaseUpdateStatusAttributes, CaseUpdateStatusRequest, CaseUpdateTitle,
    CaseUpdateTitleAttributes, CaseUpdateTitleRequest, JiraIssueCreateRequest,
    JiraIssueLinkRequest, ProjectCreate, ProjectCreateAttributes, ProjectCreateRequest,
    ProjectRelationship, ProjectRelationshipData, ProjectResourceType, ProjectUpdateRequest,
    ServiceNowTicketCreateRequest,
};

use crate::client;
use crate::config::Config;
use crate::formatter;

// ---------------------------------------------------------------------------
// Helper: build a CaseManagementAPI with bearer-token support
// ---------------------------------------------------------------------------

fn make_api(cfg: &Config) -> CaseManagementAPI {
    let dd_cfg = client::make_dd_config(cfg);
    match client::make_bearer_client(cfg) {
        Some(c) => CaseManagementAPI::with_client_and_config(dd_cfg, c),
        None => CaseManagementAPI::with_config(dd_cfg),
    }
}

// ---------------------------------------------------------------------------
// Core case operations
// ---------------------------------------------------------------------------

pub async fn search(cfg: &Config, _query: Option<String>, page_size: i64) -> Result<()> {
    let api = make_api(cfg);
    let params = SearchCasesOptionalParams::default().page_size(page_size);
    let resp = api
        .search_cases(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search cases: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn get(cfg: &Config, case_id: &str) -> Result<()> {
    let api = make_api(cfg);
    let resp = api
        .get_case(case_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get case: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn create(cfg: &Config, file: &str) -> Result<()> {
    let api = make_api(cfg);
    let body: CaseCreateRequest = crate::util::read_json_file(file)?;
    let resp = api
        .create_case(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create case: {e:?}"))?;
    formatter::output(cfg, &resp)
}

// ---------------------------------------------------------------------------
// Projects
// ---------------------------------------------------------------------------

pub async fn projects_list(cfg: &Config) -> Result<()> {
    let api = make_api(cfg);
    let resp = api
        .get_projects()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list projects: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn projects_get(cfg: &Config, project_id: &str) -> Result<()> {
    let api = make_api(cfg);
    let resp = api
        .get_project(project_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get project: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn projects_delete(cfg: &Config, project_id: &str) -> Result<()> {
    let api = make_api(cfg);
    api.delete_project(project_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete project: {e:?}"))?;
    println!("Project {project_id} deleted.");
    Ok(())
}

pub async fn projects_create(cfg: &Config, name: &str, key: &str) -> Result<()> {
    let api = make_api(cfg);
    let body = ProjectCreateRequest::new(ProjectCreate::new(
        ProjectCreateAttributes::new(key.to_string(), name.to_string()),
        ProjectResourceType::PROJECT,
    ));
    let resp = api
        .create_project(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create project: {e:?}"))?;
    formatter::output(cfg, &resp)
}

// ---------------------------------------------------------------------------
// Archive / Unarchive
// ---------------------------------------------------------------------------

pub async fn archive(cfg: &Config, case_id: &str) -> Result<()> {
    let api = make_api(cfg);
    let data = CaseEmpty::new(CaseResourceType::CASE);
    let body = CaseEmptyRequest::new(data);
    let resp = api
        .archive_case(case_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to archive case: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn unarchive(cfg: &Config, case_id: &str) -> Result<()> {
    let api = make_api(cfg);
    let data = CaseEmpty::new(CaseResourceType::CASE);
    let body = CaseEmptyRequest::new(data);
    let resp = api
        .unarchive_case(case_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to unarchive case: {e:?}"))?;
    formatter::output(cfg, &resp)
}

// ---------------------------------------------------------------------------
// Assign / Update priority / Update status
// ---------------------------------------------------------------------------

pub async fn assign(cfg: &Config, case_id: &str, user_id: &str) -> Result<()> {
    let api = make_api(cfg);
    let body = CaseAssignRequest::new(CaseAssign::new(
        CaseAssignAttributes::new(user_id.to_string()),
        CaseResourceType::CASE,
    ));
    let resp = api
        .assign_case(case_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to assign case: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn update_priority(cfg: &Config, case_id: &str, priority: &str) -> Result<()> {
    let api = make_api(cfg);
    let priority_val = match priority.to_uppercase().as_str() {
        "P1" => CasePriority::P1,
        "P2" => CasePriority::P2,
        "P3" => CasePriority::P3,
        "P4" => CasePriority::P4,
        "P5" => CasePriority::P5,
        "NOT_DEFINED" => CasePriority::NOT_DEFINED,
        _ => anyhow::bail!(
            "invalid priority: {priority} (use P1, P2, P3, P4, P5, NOT_DEFINED)"
        ),
    };
    let body = CaseUpdatePriorityRequest::new(CaseUpdatePriority::new(
        CaseUpdatePriorityAttributes::new(priority_val),
        CaseResourceType::CASE,
    ));
    let resp = api
        .update_priority(case_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update case priority: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[allow(deprecated)]
pub async fn update_status(cfg: &Config, case_id: &str, status: &str) -> Result<()> {
    let api = make_api(cfg);
    let status_val = match status.to_uppercase().as_str() {
        "OPEN" => CaseStatus::OPEN,
        "IN_PROGRESS" => CaseStatus::IN_PROGRESS,
        "CLOSED" => CaseStatus::CLOSED,
        _ => anyhow::bail!("invalid status: {status} (use OPEN, IN_PROGRESS, CLOSED)"),
    };
    let body = CaseUpdateStatusRequest::new(CaseUpdateStatus::new(
        CaseUpdateStatusAttributes::new().status(status_val),
        CaseResourceType::CASE,
    ));
    let resp = api
        .update_status(case_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update case status: {e:?}"))?;
    formatter::output(cfg, &resp)
}

// ---------------------------------------------------------------------------
// Jira integration
// ---------------------------------------------------------------------------

pub async fn jira_create_issue(cfg: &Config, case_id: &str, file: &str) -> Result<()> {
    let api = make_api(cfg);
    let body: JiraIssueCreateRequest = crate::util::read_json_file(file)?;
    api.create_case_jira_issue(case_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create Jira issue for case: {e:?}"))?;
    println!("Jira issue created for case '{case_id}'.");
    Ok(())
}

pub async fn jira_link(cfg: &Config, case_id: &str, file: &str) -> Result<()> {
    let api = make_api(cfg);
    let body: JiraIssueLinkRequest = crate::util::read_json_file(file)?;
    api.link_jira_issue_to_case(case_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to link Jira issue to case: {e:?}"))?;
    println!("Jira issue linked to case '{case_id}'.");
    Ok(())
}

pub async fn jira_unlink(cfg: &Config, case_id: &str) -> Result<()> {
    let api = make_api(cfg);
    api.unlink_jira_issue(case_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to unlink Jira issue from case: {e:?}"))?;
    println!("Jira issue unlinked from case '{case_id}'.");
    Ok(())
}

// ---------------------------------------------------------------------------
// ServiceNow integration
// ---------------------------------------------------------------------------

pub async fn servicenow_create_ticket(cfg: &Config, case_id: &str, file: &str) -> Result<()> {
    let api = make_api(cfg);
    let body: ServiceNowTicketCreateRequest = crate::util::read_json_file(file)?;
    api.create_case_service_now_ticket(case_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create ServiceNow ticket for case: {e:?}"))?;
    println!("ServiceNow ticket created for case '{case_id}'.");
    Ok(())
}

// ---------------------------------------------------------------------------
// Projects notification rules
// ---------------------------------------------------------------------------

pub async fn projects_notification_rules_list(cfg: &Config, project_id: &str) -> Result<()> {
    let api = make_api(cfg);
    let resp = api
        .get_project_notification_rules(project_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list notification rules: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn projects_notification_rules_create(
    cfg: &Config,
    project_id: &str,
    file: &str,
) -> Result<()> {
    let api = make_api(cfg);
    let body: CaseNotificationRuleCreateRequest = crate::util::read_json_file(file)?;
    let resp = api
        .create_project_notification_rule(project_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create notification rule: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn projects_notification_rules_update(
    cfg: &Config,
    project_id: &str,
    rule_id: &str,
    file: &str,
) -> Result<()> {
    let api = make_api(cfg);
    let body: CaseNotificationRuleUpdateRequest = crate::util::read_json_file(file)?;
    api.update_project_notification_rule(
        project_id.to_string(),
        rule_id.to_string(),
        body,
    )
    .await
    .map_err(|e| anyhow::anyhow!("failed to update notification rule: {e:?}"))?;
    println!("Notification rule '{rule_id}' updated.");
    Ok(())
}

pub async fn projects_notification_rules_delete(
    cfg: &Config,
    project_id: &str,
    rule_id: &str,
) -> Result<()> {
    let api = make_api(cfg);
    api.delete_project_notification_rule(project_id.to_string(), rule_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete notification rule: {e:?}"))?;
    println!("Notification rule '{rule_id}' deleted.");
    Ok(())
}

// ---------------------------------------------------------------------------
// Move case to project
// ---------------------------------------------------------------------------

pub async fn move_to_project(cfg: &Config, case_id: &str, project_id: &str) -> Result<()> {
    let api = make_api(cfg);
    let data = ProjectRelationshipData::new(project_id.to_string(), ProjectResourceType::PROJECT);
    let body = ProjectRelationship::new(data);
    let resp = api
        .move_case_to_project(case_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to move case to project: {e:?}"))?;
    formatter::output(cfg, &resp)
}

// ---------------------------------------------------------------------------
// Update case title
// ---------------------------------------------------------------------------

pub async fn update_title(cfg: &Config, case_id: &str, title: &str) -> Result<()> {
    let api = make_api(cfg);
    let attrs = CaseUpdateTitleAttributes::new(title.to_string());
    let data = CaseUpdateTitle::new(attrs, CaseResourceType::CASE);
    let body = CaseUpdateTitleRequest::new(data);
    let resp = api
        .update_case_title(case_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update case title: {e:?}"))?;
    formatter::output(cfg, &resp)
}

// ---------------------------------------------------------------------------
// Update project
// ---------------------------------------------------------------------------

pub async fn projects_update(cfg: &Config, project_id: &str, file: &str) -> Result<()> {
    let api = make_api(cfg);
    let body: ProjectUpdateRequest = crate::util::read_json_file(file)?;
    let resp = api
        .update_project(project_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update project: {e:?}"))?;
    formatter::output(cfg, &resp)
}
