from dataclasses import dataclass
from typing import List, Set


@dataclass
class ProfileFilter:
    function_prefixes: List[str]
    ignore_functions: Set[str]
