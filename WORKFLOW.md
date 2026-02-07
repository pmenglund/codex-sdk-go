# Workflow and Process Guidelines (Linear)

This document describes our development workflow, from task planning to coding, testing, and deployment. It also outlines how **ExecPlans** tie into our tracking system (Linear) to ensure every significant effort is planned and documented.

## Tracker and Task Management

- **Tracker System:** We use **Linear** for issue tracking. All work must be associated with a Linear issue:
    - Minor changes and bug fixes are usually individual issues (e.g., “Bug: fix login error”).
    - Large features or epics are tracked as **projects** or **initiatives** with multiple issues or stories.
- **Project vs Issue:** A *project/initiative* represents a large feature or project (spanning multiple days or involving multiple subtasks). An *issue* (story/task/bug) is a single unit of work (often a part of a project). By default, we create **one ExecPlan per project/initiative** to coordinate all its issues.
- **Mapping to ExecPlans:** When you start work on a project/initiative (or any task deemed complex enough), you should either locate the existing ExecPlan for it or create a new one. The ExecPlan lives in `plans/` and should list all relevant issues:
    - The ExecPlan should reference the project/initiative ID and each child issue ID (e.g., LIN-1234).
    - Every planned step in the ExecPlan should correspond to a Linear issue that is tracked. Conversely, every Linear issue under that project/initiative should be reflected in the ExecPlan’s **Progress** or **Plan** section. This one-to-one mapping ensures no work is done off the books.

**When is an ExecPlan required?** – Use an ExecPlan for any work that:
- Is a **project/initiative** in Linear (default rule: 1 ExecPlan per project/initiative).
- Involves a **significant refactor** or architectural change affecting multiple components.
- Introduces uncertainty or research (spiking on a solution, multiple unknowns).
- Will take more than a day or two of effort, or requires multiple PRs to complete.
- *Optional:* For smaller issues, an ExecPlan is usually not needed. However, if the agent or developer feels a plan would help (e.g. the task is tricky or the solution space is unclear), they can create a lightweight ExecPlan or checklist even if not a project.

If a task is **not** complex (straightforward bug fix or minor tweak), you may proceed without an ExecPlan, but still follow coding and testing guidelines. Use judgment: *“When in doubt, make a plan.”* It’s better to have an ExecPlan that turns out simple than to dive in and get lost.

## Linear MCP Server Usage

Use the **Linear MCP server** to fetch, verify, and update issue context throughout the workflow:

- **Lookup:** Pull the Linear issue details (title, description, status, assignee, project) at the start of work so the plan is grounded in the source of truth.
- **Relationship mapping:** Confirm the parent project/initiative and list all child issues when building or updating an ExecPlan.
- **Status updates:** Move issues through states (e.g., Backlog → In Progress → Done) as work progresses.
- **Notes and links:** Attach relevant commit/PR links or notes to the Linear issue when appropriate.

If Linear MCP is unavailable, proceed using the best available context and note the gap in the ExecPlan or PR description.

## Planning & Execution Flow

1. **Pick up a Task:** The AI agent should start by identifying the Linear issue it’s working on (usually provided in the prompt or context by the user). Verify whether it is part of a project/initiative.
2. **Ensure Context:** If it’s a project/initiative or complex task, check if an ExecPlan exists in `plans/`.
    - If yes, open it and review the plan and progress.
    - If no and the task meets the ExecPlan criteria, **create a new ExecPlan** file (see PLANS.md for the template and required sections). Name it clearly (e.g., `plans/LIN-<PROJECT-ID>.md` or a descriptive slug). Populate it following the standard and include references to all child issues (use Linear MCP to list them).
3. **Plan Approval:** (If in team setting) Optionally, get the ExecPlan reviewed by a lead or the user before coding. In AI’s case, it may proceed autonomously but should have high confidence in the plan (or get user confirmation if unsure).
4. **Implement in Steps:** With or without an ExecPlan, the agent should break work into small steps:
    - If using ExecPlan: follow the “Plan of Work” and “Concrete Steps” in the plan. Update the **Progress** checklist as each sub-task is completed. Mark tasks done (with timestamp) in both the plan and in Linear (move issue to Done or attach commit messages).
    - If no ExecPlan (small task): the agent should still outline the solution approach briefly (either mentally or by a short comment) and proceed in logical increments (similar to plan steps).
5. **Frequent Commits:** Commit after each meaningful step or sub-task:
    - Include the Linear issue ID in commit messages (e.g., “LIN-456: add null check” or “Fix input validation (LIN-456)”). This auto-links commits to issues.
    - Ensure each commit compiles and passes tests (see Quality and Testing below).
    - If a step involves a design decision (e.g., choosing one library over another), note it in the ExecPlan’s Decision Log as well as possibly in the commit message.
6. **Testing Each Step:** Run tests locally after each major code change. Don’t wait until the end to test – catch issues early. If adding a new feature, ideally write tests **before** or alongside the code for that feature (as per TDD guidance).
7. **Progress Updates:** Keep Linear updated:
    - Move issues to In Progress / Done as appropriate.
    - If working with an ExecPlan, sync its Progress section with reality. For example, when a sub-task is completed, mark it `[x]` with a timestamp and perhaps add a brief note if needed.
    - If you discover new tasks during implementation (scope changes or additional bugs), add them to Linear and to the ExecPlan (in Progress or as new planned steps), so nothing falls through the cracks.

## Quality Gates (Validation & Success Criteria)

- **Code Review:** Open a Pull Request (PR) for the changes. The PR should ideally be tied to the project/initiative or parent issue. In the PR description:
    - Summarize what the change accomplishes and **how to test it** (if not obvious).
    - If an ExecPlan was used, mention it (e.g., “Implemented according to plan in `plans/LIN-1234.md`”) and ensure all plan items are completed.
    - List the Linear issues addressed by this PR (e.g., “Closes LIN-456, LIN-457”).
    - If not all issues of the project are done (e.g., splitting into multiple PRs), note what remains and that the plan will continue.
- **Continuous Integration (CI):** Our CI pipeline runs automatically on each PR:
    - Code formatting and lint checks (must pass).
    - All unit and integration tests (must pass).
    - (If applicable) Security or type checks, etc.
    - The agent should ensure these all pass *before* marking a PR ready. It can run the same commands locally (see below).
- **Manual QA (if any):** For certain features, a team member might do a manual test in a staging environment. The ExecPlan’s **Validation** section (and APP.md’s instructions) should outline how to run the app and what to verify. The agent should have already validated the acceptance criteria as per the plan.

**Definition of Done:** A task/feature is considered done when:
- Its code is written and meets all acceptance criteria (as per tests and validation steps).
- All **mandatory tests** are written and passing.
- Documentation (code comments, relevant markdown docs) is updated.
- The ExecPlan (if one was used) is fully updated – Progress 100%, Decision Log complete, Outcomes written.
- The change is merged into the main branch (after code review approval).
- The Linear issue(s) are closed/resolved.

## Branching and Release Workflow

*(Adjust this section to your Git branching model.)* For example:
- We follow a simple model: feature work is done in feature branches named after the task (e.g., `feat/LIN-1234-user-profile`). The agent can create a branch when starting on a project or issue.
- Once work is done (or at a good stopping point), open a PR from that branch to `main` (or `develop` if using Git Flow).
- Multiple PRs can reference the same ExecPlan/project if we deliver in increments; ensure the ExecPlan covers the full project and is updated as each PR completes part of it.
- Releases are cut from `main` periodically. Ensure that all merged code has associated closed issues and (if applicable) completed ExecPlans.

## Automation & Compliance

To enforce this workflow:
- We use **branch protection and CI** to block merging if checks fail or if PR lacks linked issues.
- Optionally, a bot or script may verify that for any PR referencing a project, a corresponding ExecPlan file exists in `plans/`. (E.g., if PR title has LIN-1234, ensure `plans/LIN-1234.md` is in the diff or already in repo.) This catches if someone forgot to make a plan.
- We might introduce a checklist in PR templates reminding contributors (and the AI) to follow these steps (e.g., “- [ ] Tests passing, - [ ] ExecPlan updated, - [ ] Issue linked”).
- The AI assistant is expected to self-check compliance: after finishing, it should confirm to the user that all process steps were followed (or flag if something like an ExecPlan was skipped).

By adhering to this WORKFLOW, we ensure that AI-assisted development is transparent and trackable. Every feature has a paper trail (via Linear and ExecPlans), and nothing significant happens without oversight. This maintains **accountability** and prevents “rogue” changes outside the planning process.
