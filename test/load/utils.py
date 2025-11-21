import random
import string
from typing import Dict, List


def make_team(name: str, user_count: int):
    members = []
    for idx in range(1, user_count + 1):
        members.append(
            {
                "user_id": f"{name}-user-{idx}",
                "username": f"user-{idx}",
                "is_active": True,
            }
        )
    return {"team_name": name, "members": members}


def make_prs_for_user(author: str, count: int):
    prs = []
    for idx in range(1, count + 1):
        prs.append(
            {
                "pull_request_id": f"{author}-pr-{idx}-{_rand_suffix()}",
                "pull_request_name": f"Feature #{idx}",
                "author_id": author,
            }
        )
    return prs


def _rand_suffix(length: int = 4) -> str:
    return "".join(random.choices(string.ascii_lowercase + string.digits, k=length))
