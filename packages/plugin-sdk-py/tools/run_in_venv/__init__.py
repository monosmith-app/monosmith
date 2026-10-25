"""Runs a Python module with the workspace's virtual environment (`python -m tools.run_in_venv.run_in_venv_handler`)."""

from .venv_interpreter_algorithm import find_venv_interpreter

__all__ = ["find_venv_interpreter"]
