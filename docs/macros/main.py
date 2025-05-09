"""
Basic example of a Mkdocs-macros module
"""

from functools import cache
from typing import Any, Dict, Literal

import requests
from mkdocs_macros.plugin import MacrosPlugin


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
        "tag_name": "v0.19.1",
    }


@cache
def latest_binary(
    platform: Literal["linux", "windows"],
    version: str,
    arch: str = "amd64",
):
    suffix = ".exe" if platform == "windows" else ""
    return f"situation-{version}-{arch}-{platform}{suffix}"


def define_env(env: MacrosPlugin):
    """
    This is the hook for defining variables, macros and filters

    - variables: the dictionary that contains the environment variables
    - macro: a decorator function, to declare a macro.
    - filter: a function with one of more arguments,
        used to perform a transformation
    """
    img_dir = "../img"

    env.variables["github_repo"] = "https://github.com/situation-sh/situation"
    env.variables["windows_ok"] = (
        f"""![windows]({img_dir}/windows.svg){{: class="tag" title="Windows" }}"""
    )
    env.variables["linux_ok"] = (
        f"""![linux]({img_dir}/linux.svg){{: class="tag" title="Linux"}}"""
    )
    env.variables["root_required"] = (
        f"""![root]({img_dir}/root.svg){{: class="tag" title="Root required" }}"""
    )
    env.variables["latest_tag"] = latest_release().get("tag_name", "")
    env.variables["latest_version"] = env.variables["latest_tag"].removeprefix("v")
    env.variables["latest_linux_binary"] = latest_binary(
        "linux",
        env.variables["latest_version"],
    )
    env.variables["latest_windows_binary"] = latest_binary(
        "windows",
        env.variables["latest_version"],
    )

    env.variables["windows_icon_src"] = f"{img_dir}/windows.svg"
    env.variables["linux_icon_src"] = f"{img_dir}/linux.svg"
    env.variables["root_required_icon_src"] = f"{img_dir}/root.svg"
