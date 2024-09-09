from base_test import BaseTest


class UrlTest(BaseTest):
    def __init__(self):
        super().__init__()

    def _shorUrlGeneration(self, website: str) -> str:
        # 测试短链生成
        url = "/urls/shorten"
        return self._req("POST", url, json={"url": website})

    def testUrlGenAndVisit(self):
        urls = [
            "https://www.baidu.com"
        ]

        res: dict[str, str] = {url: self._shorUrlGeneration(url) for url in urls}


if __name__ == '__main__':
    short_url = UrlTest().run()
