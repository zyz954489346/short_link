import hmac
import base64
import hashlib
import inspect
import requests
import urllib.parse
from pprint import pp


class BaseTest:
    def __init__(self):
        self.domain: str = "http://localhost:8080"
        self.key: str = "p79KKyJTgfG2snUs"
        self.secret: str = "s6NmXR0E8pPd23KT"

    def _buildSignature(self, params: dict) -> str:
        """
        签名构建
        :param params: 参数
        :return: 签名
        """

        # key 升序
        sorted_params = dict(sorted(params.items()))
        # 拼接
        query_str = "&".join([f"{k}={v}" for k, v in sorted_params.items()])
        # 创建一个新的 HMAC 使用 sha256
        h = hmac.new(self.secret.encode(), query_str.encode(), hashlib.sha256)
        # Base64 加密
        base64_str = base64.b64encode(h.digest()).decode()
        # url Encode
        return urllib.parse.quote(base64_str)

    def _generateParams(self, params: dict) -> dict:
        """
        请求参数构建
        :param params: 参数
        :return: 全部参数
        """

        params["key"] = self.key
        params["sign"] = self._buildSignature(params)
        return params

    def _req(self, method: str, url: str, **kwargs) -> any:
        api_url = self.domain + url
        params = self._generateParams(kwargs["json"] if "json" in kwargs else {})
        print(method, api_url, params)
        response = requests.request(method, api_url, json=params)
        response.raise_for_status()
        result = response.json()

        if result["code"] != 0:
            raise Exception(result["message"])
        else:
            pp(result)

        return result["data"]

    def run(self):
        # 获取所以可调用方法
        for func_name, func in inspect.getmembers(self, predicate=callable):
            if not func_name.startswith("test"):
                continue
            func()
