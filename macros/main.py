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
    # env.variables[
    #     "windows_ok"
    # ] = "![windows](https://img.shields.io/badge/windows-xxx?style=flat-square&logo=windows&labelColor=%230078D6&color=%230078D6)"
    env.variables["windows_ok"] = (
        """![windows](https://img.shields.io/badge/windows-xxx?style=flat-square&labelColor=%230078D6&color=%230078D6&logoColor=%23FFFFFF&logo=data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI1NiIgaGVpZ2h0PSI1NiIgdmlld0JveD0iMCAwIDU2IDU2Ij48cGF0aCBmaWxsPSJ3aGl0ZSIgZD0ibTUgMTEuNTMzbDE4Ljc5OS0yLjU2bC4wMDggMTguMTMzbC0xOC43OS4xMDd6bTE4Ljc5IDE3LjY2MmwuMDE0IDE4LjE0OWwtMTguNzktMi41ODRWMjkuMDczem0yLjI3OS0yMC41NTdMNTAuOTk0IDV2MjEuODc1bC0yNC45MjUuMTk4ek01MSAyOS4zNjZsLS4wMDYgMjEuNzc2bC0yNC45MjUtMy41MThsLS4wMzUtMTguM3oiLz48L3N2Zz4=)"""
    )
    env.variables["linux_ok"] = (
        "![linux](https://img.shields.io/badge/linux-a?style=flat-square&logo=linux&logoColor=%23000000&labelColor=%23FCC624&color=%23FCC624)"
    )
    env.variables["root_required"] = (
        "![root](https://img.shields.io/badge/root_required-a?style=flat-square&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgLTk2MCA5NjAgOTYwIiB3aWR0aD0iNDgiPjxwYXRoIGZpbGw9IndoaXRlIiBkPSJNMjM2LjE3NC03Ni40MTN2LTE3MS4xOTZxLTM1LjA0NC0xNi4yMzktNjQuNDI0LTQzLjQ3OC0yOS4zOC0yNy4yMzktNTAuNjItNjIuOTc4LTIxLjIzOS0zNS43MzktMzIuOTc4LTc3Ljk3OFE3Ni40MTMtNDc0LjI4MyA3Ni40MTMtNTIwcTAtMTU5LjY3NCAxMTIuOTktMjYxLjYzIDExMi45ODktMTAxLjk1NyAyOTAuNjMtMTAxLjk1N1Q3NzAuNjMtNzgxLjYzUTg4My41ODctNjc5LjY3NCA4ODMuNTg3LTUyMHEwIDQ1LjcxNy0xMS43MzkgODcuOTU3LTExLjczOSA0Mi4yMzktMzIuOTc4IDc3Ljk3OC0yMS4yNCAzNS43MzktNTAuNjIgNjIuOTc4LTI5LjM4IDI3LjIzOS02NC40MjQgNDMuNDc4djE3MS4xOTZIMjM2LjE3NFptNjguMzctNjQuNzgzaDUyLjgyNnY2NC43ODNoNjEuNzUzdi02NC43ODNoMTIxLjc1NHY2NC43ODNoNjEuNzUzdi02NC43ODNoNTIuODI2di0xNDdxMzctMTIuMTk1IDY2LjUtMzQuNDM0IDI5LjUtMjIuMjQgNTAuMzgxLTUyLjM4MSAyMC44OC0zMC4xNDEgMzItNjYuOTQzIDExLjExOS0zNi44MDIgMTEuMTE5LTc3Ljg5MiAwLTEzMi4xMTQtOTIuNzgzLTIxMy44NjItOTIuNzgzLTgxLjc0OC0yNDIuNjA5LTgxLjc0OC0xNDkuODI1IDAtMjQyLjU1MyA4MS43NTctOTIuNzI4IDgxLjc1OC05Mi43MjggMjEzLjg4NyAwIDQxLjA5NSAxMSA3Ny44NzcgMTEgMzYuNzgzIDMxLjg4IDY2LjkyNCAyMC44ODEgMzAuMTQxIDUwLjM4MSA1Mi4zODEgMjkuNSAyMi4yMzkgNjYuNSAzNC40MzR2MTQ3Wm0xMTEuODY5LTE5Mi4xMDhoMTI3LjE3NEw0ODAtNDYwLjQ3OGwtNjMuNTg3IDEyNy4xNzRabS03Ni4yODQtMTI1LjAyMnEyOS44MjggMCA1MC44MDYtMjEuMTA3IDIwLjk3OC0yMS4xMDcgMjAuOTc4LTUwLjY5NiAwLTI5LjgyOC0yMS4wMzctNTAuODA2LTIxLjAzNy0yMC45NzgtNTAuOTM1LTIwLjk3OC0yOS42NTggMC01MC42MzcgMjEuMDM3LTIwLjk3OCAyMS4wMzctMjAuOTc4IDUwLjkzNSAwIDI5LjY1OCAyMS4xMDcgNTAuNjM3IDIxLjEwNyAyMC45NzggNTAuNjk2IDIwLjk3OFptMjgwIDBxMjkuODI4IDAgNTAuODA2LTIxLjEwNyAyMC45NzgtMjEuMTA3IDIwLjk3OC01MC42OTYgMC0yOS44MjgtMjEuMDM3LTUwLjgwNi0yMS4wMzctMjAuOTc4LTUwLjkzNS0yMC45NzgtMjkuNjU4IDAtNTAuNjM3IDIxLjAzNy0yMC45NzggMjEuMDM3LTIwLjk3OCA1MC45MzUgMCAyOS42NTggMjEuMTA3IDUwLjYzNyAyMS4xMDcgMjAuOTc4IDUwLjY5NiAyMC45NzhabS0zMTUuNTg1IDMxNy4xM3YtMTQ3cS0zNy0xMi4xOTUtNjYuNS0zNC40MzQtMjkuNS0yMi4yNC01MC4zODEtNTIuMzgxLTIwLjg4LTMwLjE0MS0zMS44OC02Ni45NDN0LTExLTc3Ljg5MnEwLTEzMi4xMTQgOTIuNzM3LTIxMy44NjIgOTIuNzM4LTgxLjc0OCAyNDIuNDg5LTgxLjc0OCAxNDkuNzUyIDAgMjQyLjYgODEuNzU3IDkyLjg0NyA4MS43NTggOTIuODQ3IDIxMy44ODcgMCA0MS4wOTUtMTEuMTE5IDc3Ljg3Ny0xMS4xMiAzNi43ODMtMzIgNjYuOTI0LTIwLjg4MSAzMC4xNDEtNTAuMzgxIDUyLjM4MS0yOS41IDIyLjIzOS02Ni41IDM0LjQzNHYxNDdINjAyLjYzdi02MGgtNjEuNzUzdjYwSDQxOS4xMjN2LTYwSDM1Ny4zN3Y2MGgtNTIuODI2WiIvPjwvc3ZnPgo%3D&logoColor=%23FFFFFF&labelColor=%23E01F27&color=%23E01F27)"
    )

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
            "tag_name": "v0.14.1",
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
