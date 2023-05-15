"""
Basic example of a Mkdocs-macros module
"""

from functools import cache
from typing import Any, Dict, Literal

import requests
from mkdocs_macros.plugin import MacrosPlugin


def define_env(env: MacrosPlugin):
    """
    This is the hook for defining variables, macros and filters

    - variables: the dictionary that contains the environment variables
    - macro: a decorator function, to declare a macro.
    - filter: a function with one of more arguments,
        used to perform a transformation
    """

    env.variables["github_repo"] = "https://github.com/situation-sh/situation"

    @cache
    def latest_release() -> Dict[str, Any]:
        response = requests.get(
            "https://api.github.com/repos/situation-sh/situation/releases/latest",
            headers={
                "Accept": "application/vnd.github+json",
                "X-GitHub-Api-Version": "2022-11-28",
            },
            timeout=1000,
        )
        if response.status_code == 200:
            return response.json()
        print("error:", response.content)
        return {
            "tag_name": "v0.13.8",
        }

    @env.macro
    def latest_tag() -> str:
        data = latest_release()
        return data.get("tag_name", "")

    @env.macro
    def latest_version():
        tag = latest_tag()
        return tag.removeprefix("v")

    @env.macro
    def latest_binary(
        platform: Literal["linux", "windows"],
        arch: str = "amd64",
    ):
        suffix = ".exe" if platform == "windows" else ""
        return f"situation-{latest_version()}-{arch}-{platform}{suffix}"
