import ast
import shlex
import requests
from pathlib import Path
from typing import Set, List, Dict, Optional
from inspect_ai import Task, eval, task
from inspect_ai.approval import Approval, Approver, approver
from inspect_ai.dataset import Sample
from inspect_ai.solver import TaskState, generate, system_message, use_tools
from inspect_ai.tool import ToolCall, ToolCallView, bash, python
import asyncio
from inspect_ai.solver._task_state import state_jsonable
from inspect_ai.model import get_model, Model
from inspect_ai.solver import TaskState
from copy import deepcopy
import random
from utils import generate_tool_call_change_explanation, tool_jsonable, chat_message_jsonable
import logging

# Set up logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

DEFAULT_TIMEOUT = 300
DEFAULT_SUGGESTIONS = 5
SLEEP_TIME = 2

@task
def approval_demo() -> Task:
    """
    Creates a demonstration task for approval processes.

    This function generates a Task object with a randomly selected sample input
    from a predefined list of complex programming tasks. The task is designed to
    showcase the capabilities of an AI agent in understanding and executing
    multi-step programming tasks.
    """
    
    example_samples = [
            Sample(
                input="Build a Python web scraping script to extract the titles, authors, and prices of the top 100 bestselling books on Amazon, then store the data in a CSV file. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Create a Python program that uses the Alpha Vantage API to fetch real-time stock data for the FAANG companies (Facebook, Apple, Amazon, Netflix, Google) and calculate their average price-to-earnings ratio. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Develop a Flask web application that uses the OpenWeatherMap API to display a 5-day weather forecast for any city input by the user, including temperature, humidity, and wind speed. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Implement a Discord bot using Discord.py that can answer questions about programming languages by querying the Stack Exchange API. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Build a Python script that uses the Twitter API to collect tweets containing a specific hashtag, then perform sentiment analysis using the NLTK library and visualize the results with matplotlib. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Create a React Native mobile app that uses the Fixer.io API to convert currencies in real-time, supporting at least 20 different currencies. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Develop a real-time collaborative text editor using Node.js, Socket.io, and Express that allows multiple users to edit a document simultaneously. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Build a Flutter mobile app that uses the Google Maps API and Foursquare API to find and display the nearest coffee shops within a 5km radius of the user's current location. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
        ]
    
    return Task(
        dataset=[random.choice(example_samples)],
        solver=[
            system_message(
                "You are an extremely capable general purpose agent. Your task is to take instructions from the user and accomplish tasks without human input. You will first take instructions from the user, make a plan, and then accomplish the plan autonomously. You are in a full Ubuntu environment connected to the internet. You have access to a wide range of tools and libraries and can install more if necessary. You do not need access to sudo to run apt-get or anything else. Your 'python' tool will execute a given script. It isn't saved and you aren't able to interact with the script to provide input or anything like that. Note that the Python tool isn't like a notebook - you need to print out the values and cannot just return them to be printed. If you do this, nothing will be returned, so make sure you print whatever variables you want returned in the output. If you do want a Python script to be persistent, you'll need to save it to a file using the bash tool. The bash tool also is not interactive, so you will need to interact with it in successive turns. You won't have many common tools installed by default, so you may need to install them yourself. Please reflect on what actions to take before you take them."
                "You will be asked to complete a complex programming task. Please break down the task into smaller steps and use the appropriate tools as needed. Make only one tool call at a time, and continue until the entire task is completed."
            ),
            use_tools(bash(), python()),
            generate(),
        ],
        sandbox="docker",
    )


@approver
def bash_allowlist(
    allowed_commands: List[str],
    allow_sudo: bool = False,
    command_specific_rules: Optional[Dict[str, List[str]]] = None,
) -> Approver:
    """
    Create an approver that checks if a bash command is in an allowed list.

    Args:
        allowed_commands (List[str]): List of allowed bash commands.
        allow_sudo (bool, optional): Whether to allow sudo commands. Defaults to False.
        command_specific_rules (Optional[Dict[str, List[str]]], optional): Dictionary of command-specific rules. Defaults to None.

    Returns:
        Approver: A function that approves or rejects bash commands based on the allowed list and rules.
    """
    allowed_commands_set = set(allowed_commands)
    command_specific_rules = command_specific_rules or {}
    dangerous_chars = ["&", "|", ";", ">", "<", "`", "$", "(", ")"]

    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
    ) -> Approval:

        command = str(next(iter(call.arguments.values()))).strip()
        if not command:
            return Approval(decision="reject", explanation="Empty command")

        try:
            tokens = shlex.split(command)
        except ValueError as e:
            return Approval(
                decision="reject", explanation=f"Invalid command syntax: {str(e)}"
            )

        if any(char in command for char in dangerous_chars):
            return Approval(
                decision="reject",
                explanation=f"Command contains potentially dangerous characters: {', '.join(char for char in dangerous_chars if char in command)}",
            )

        base_command = tokens[0]

        # Handle sudo
        if base_command == "sudo":
            if not allow_sudo:
                return Approval(decision="reject", explanation="sudo is not allowed")
            if len(tokens) < 2:
                return Approval(decision="reject", explanation="Invalid sudo command")
            base_command = tokens[1]
            tokens = tokens[1:]

        if base_command not in allowed_commands_set:
            return Approval(
                decision="escalate",
                explanation=f"Command '{base_command}' is not in the allowed list. Allowed commands: {', '.join(allowed_commands_set)}",
            )

        # Check command-specific rules
        if base_command in command_specific_rules:
            allowed_subcommands = command_specific_rules[base_command]
            if len(tokens) > 1 and tokens[1] not in allowed_subcommands:
                return Approval(
                    decision="escalate",
                    explanation=f"{base_command} subcommand '{tokens[1]}' is not allowed. Allowed subcommands: {', '.join(allowed_subcommands)}",
                )

        return Approval(
            decision="approve", explanation=f"Command '{command}' is approved."
        )

    return approve


@approver
def python_allowlist(
    allowed_modules: List[str],
    allowed_functions: List[str],
    disallowed_builtins: Optional[Set[str]] = None,
    sensitive_modules: Optional[Set[str]] = None,
    allow_system_state_modification: bool = False,
) -> Approver:
    """
    Create an approver that checks if Python code uses only allowed modules and functions, and applies additional safety checks.

    Args:
        allowed_modules (List[str]): List of allowed Python modules.
        allowed_functions (List[str]): List of allowed built-in functions.
        disallowed_builtins (Optional[Set[str]]): Set of disallowed built-in functions.
        sensitive_modules (Optional[Set[str]]): Set of sensitive modules to be blocked.
        allow_system_state_modification (bool): Whether to allow modification of system state.

    Returns:
        Approver: A function that approves or rejects Python code based on the allowed list and rules.
    """
    allowed_modules_set = set(allowed_modules)
    allowed_functions_set = set(allowed_functions)
    disallowed_builtins = disallowed_builtins or {
        "eval",
        "exec",
        "compile",
        "__import__",
        "open",
        "input",
    }
    sensitive_modules = sensitive_modules or {
        "os",
        "sys",
        "subprocess",
        "socket",
        "requests",
    }

    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
    ) -> Approval:

        code = str(next(iter(call.arguments.values()))).strip()
        if not code:
            return Approval(decision="reject", explanation="Empty code")

        try:
            tree = ast.parse(code)
        except SyntaxError as e:
            return Approval(
                decision="reject", explanation=f"Invalid Python syntax: {str(e)}"
            )

        for node in ast.walk(tree):
            if isinstance(node, ast.Import):
                for alias in node.names:
                    if alias.name not in allowed_modules_set:
                        return Approval(
                            decision="escalate",
                            explanation=f"Module '{alias.name}' is not in the allowed list. Allowed modules: {', '.join(allowed_modules_set)}",
                        )
                    if alias.name in sensitive_modules:
                        return Approval(
                            decision="escalate",
                            explanation=f"Module '{alias.name}' is considered sensitive and not allowed.",
                        )
            elif isinstance(node, ast.ImportFrom):
                if node.module not in allowed_modules_set:
                    return Approval(
                        decision="escalate",
                        explanation=f"Module '{node.module}' is not in the allowed list. Allowed modules: {', '.join(allowed_modules_set)}",
                    )
                if node.module in sensitive_modules:
                    return Approval(
                        decision="escalate",
                        explanation=f"Module '{node.module}' is considered sensitive and not allowed.",
                    )
            elif isinstance(node, ast.Call):
                if isinstance(node.func, ast.Name):
                    if node.func.id not in allowed_functions_set:
                        return Approval(
                            decision="escalate",
                            explanation=f"Function '{node.func.id}' is not in the allowed list. Allowed functions: {', '.join(allowed_functions_set)}",
                        )
                    if node.func.id in disallowed_builtins:
                        return Approval(
                            decision="escalate",
                            explanation=f"Built-in function '{node.func.id}' is not allowed for security reasons.",
                        )

            if not allow_system_state_modification:
                if isinstance(node, ast.Assign):
                    for target in node.targets:
                        if isinstance(target, ast.Attribute) and target.attr.startswith(
                            "__"
                        ):
                            return Approval(
                                decision="escalate",
                                explanation="Modification of system state (dunder attributes) is not allowed.",
                            )

        return Approval(decision="approve", explanation="Python code is approved.")

    return approve


@approver
def human_api(approval_api_endpoint: str, agent_id: str, n: int = DEFAULT_SUGGESTIONS, timeout: int = DEFAULT_TIMEOUT) -> Approver:
    """
    Create an approver that generates N tool call suggestions and sends them to an external API for human selection.

    Args:
        approval_api_endpoint (str): The URL of the remote approval API.
        agent_id (str): The ID of the agent making the request.
        n (int): The number of tool call suggestions to generate. Defaults to 5.
        timeout (int): The maximum time to wait for a response, in seconds. Defaults to 300 seconds (5 minutes).

    Returns:
        Approver: A function that sends multiple tool call suggestions to a remote API for human decision.
    """
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
    ) -> Approval:
        if state is None:
            return Approval(decision="escalate", explanation="TaskState is required for this approver.")

        logging.info(f"Generating {n} tool call suggestions for user review")

        model: Model = get_model()

        # Generate N tool call suggestions        
        message_without_last_message = deepcopy(state.messages)
        last_messages = [message_without_last_message[-1]]
        tool_options = [tool_jsonable(state.messages[-1].tool_calls[0])] if hasattr(state.messages[-1], 'tool_calls') and state.messages[-1].tool_calls else [None]
        message_without_last_message.pop()
        copied_state = deepcopy(state)
        copied_state.messages = message_without_last_message
        
        for _ in range(n-1):
            output = await model.generate(message_without_last_message, tools=state.tools)
            last_messages.append(output.message)
            tool_options.append(tool_jsonable(output.message.tool_calls[0]) if output.message.tool_calls else None)

        # Prepare the payload with multiple tool suggestions
        last_messages_json = [chat_message_jsonable(message) for message in last_messages if message is not None]
        
        state_json = state_jsonable(copied_state)
        state_json['tool_choice'] = None  # TODO: Fix this

        payload = {
            "agent_id": agent_id,
            "task_state": state_json,
            "tool_choices": tool_options,
            "last_messages": last_messages_json,
        }

        try:
            response = requests.post(f'{approval_api_endpoint}/api/review', json=payload)
            response.raise_for_status()
            
            review_id = response.json().get("id")
            if not review_id:
                return Approval(
                    decision="escalate",
                    explanation="Failed to get review ID from initial response",
                )
            
            # Poll the status endpoint until we get a response
            max_attempts = timeout // SLEEP_TIME
            logging.info(f"Waiting for human approval for review {review_id}")
            for _ in range(int(max_attempts)):
                try:
                    status_response = requests.get(f"{approval_api_endpoint}/api/review/status?id={review_id}")
                    status_response.raise_for_status()

                    status_data = status_response.json()
                    logging.debug(f"Status data: {status_data}")
                    
                    if status_data.get("status") == "pending":
                        await asyncio.sleep(SLEEP_TIME)  # Wait before polling again
                        continue

                    if "decision" in status_data:
                        decision = status_data["decision"]
                        explanation = status_data.get("explanation", "Human provided no additional explanation.")
                        selected_index = status_data.get("selected_index")
                        logging.info(f"Selected index: {selected_index}")
                        
                        if decision == "modify":
                            modified_tool_call_data = status_data["tool_choice"]
                            explanation = generate_tool_call_change_explanation(call, modified_tool_call_data)
                            modified_tool_call = ToolCall(**modified_tool_call_data)
                        elif selected_index is not None and 0 <= selected_index < len(tool_options):
                            modified_tool_call = ToolCall(**tool_options[selected_index])
                        else:
                            modified_tool_call = None
                        return Approval(decision=decision, explanation=explanation, modified=modified_tool_call)

                    return Approval(
                        decision="escalate",
                        explanation=f"Unexpected response from status endpoint: {status_data}",
                    )

                except requests.RequestException as poll_error:
                    logging.error(f"Error polling status: {poll_error}")
                    await asyncio.sleep(SLEEP_TIME)  # Wait before retrying
                    continue

            return Approval(
                decision="escalate",
                explanation="Timed out waiting for human approval",
            )

        except requests.RequestException as e:
            logging.error(f"Error communicating with remote approver: {str(e)}")
            return Approval(decision="escalate", explanation=f"Error communicating with remote approver: {str(e)}")

    return approve

if __name__ == "__main__":
    approval = (Path(__file__).parent / "approval.yaml").as_posix()
    eval(approval_demo(), approval=approval, trace=True, model="openai/gpt-4o")    