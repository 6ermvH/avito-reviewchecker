import os
import random

from locust import HttpUser, between, task

from utils import make_prs_for_user, make_team

COUNT_USERS_PER_TEAM = 20
COUNT_PRS_PER_USER = 5
WAIT_MIN = 0.05
WAIT_MAX = 0.2

TEAMS = []
USERS = []
PRS = []

IDX = 1

class ServiceUser(HttpUser):
    wait_time = between(WAIT_MIN, WAIT_MAX)

    def on_start(self):
        self._seed_environment()

    def _seed_environment(self):
        global IDX
        team_name = f"load-team-{IDX}"
        IDX += 1
        payload = make_team(team_name, COUNT_USERS_PER_TEAM)
        with self.client.post("/team/add", json=payload, catch_response=True) as resp:
            if resp.status_code not in (200, 201):
                resp.failure(f"seed team failed: {resp.status_code} {resp.text}")
                return
            resp.success()
            TEAMS.append(team_name)

        for member in payload["members"]:
            user_id = member["user_id"]
            USERS.append(user_id)


    @task
    def create_pr(self):
        self._task_create_pr()

    @task
    def merge_pr(self):
        self._task_merge_pr()

    @task
    def reassign_pr(self):
        self._task_reassign_pr()

    def _task_create_pr(self):
        author = random.choice(USERS)
        payload = make_prs_for_user(author, random.randint(1, COUNT_PRS_PER_USER))
        for pr in payload:
            with self.client.post("/pullRequest/create", json=pr, catch_response=True) as resp:
                if resp.status_code == 201:
                    data = resp.json()
                    reviewers = data.get("pr", {}).get("assigned_reviewers")
                    PRS.append({
                        "id": pr["pull_request_id"],
                        "author": author,
                        "reviewers": reviewers,
                    })
                    resp.success()

    def _task_merge_pr(self):
        if len(PRS) == 0:
            return
        pr = random.choice(PRS)
        resp = self.client.post(
            "/pullRequest/merge",
            json={"pull_request_id": pr["id"]},
        )
        if resp.status_code == 200:
            PRS.remove(pr)

    def _task_reassign_pr(self):
        candidates = [pr for pr in PRS if pr["reviewers"]]
        if not candidates:
            return
        pr = random.choice(candidates)
        old_user = random.choice(pr["reviewers"])
        with self.client.post(
            "/pullRequest/reassign",
            json={
                "pull_request_id": pr["id"],
                "old_user_id": old_user,
            },
            catch_response=True,
        ) as resp:
            if resp.status_code == 200:
                data = resp.json()
                reviewers = data.get("pr", {}).get("assigned_reviewers")
                if reviewers:
                    pr["reviewers"] = reviewers
                else:
                    replaced_by = data.get("replaced_by")
                    if replaced_by:
                        try:
                            pr["reviewers"].remove(old_user)
                        except ValueError:
                            pass
                        if replaced_by not in pr["reviewers"]:
                            pr["reviewers"].append(replaced_by)
                resp.success()
