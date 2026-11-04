"""奇门工具层表结构错误。"""

from app.natal.tools.errors import LikiToolError


class TableError(LikiToolError, ValueError):
    """事象表或解释表结构错误。"""
