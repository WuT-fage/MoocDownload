# -*- coding: utf-8 -*-
"""
@File    :   loginWeb.py
@Contact :   jyj345559953@qq.com
@Author  :   Esword
"""

import re
import requests
from flask import Flask, Response, request
from requests.utils import dict_from_cookiejar

session = requests.session()
session.headers = {
    'referer': 'https://www.icourse163.org/',
    'sec-fetch-dest': 'empty',
    'sec-fetch-mode': 'cors',
    'sec-fetch-site': 'same-origin',
    'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.190 Safari/537.36',
}

app = Flask(__name__)

pollKey = ""


@app.get("/qrcode")
def QrcodePoll():
    global pollKey
    session.get("https://www.icourse163.org/")
    url = "https://www.icourse163.org/logonByQRCode/code.do?width=182&height=182"
    response = session.get(url=url)
    try:
        dic = response.json()
        codeUrl = dic["result"]["codeUrl"]
        pollKey = dic["result"]["pollKey"]
        return codeUrl
    except:
        return "登录失败"


@app.get("/pollKey")
def pollKey():
    global pollKey
    url = f"https://www.icourse163.org/logonByQRCode/poll.do?pollKey={pollKey}"
    response = session.get(url=url)
    Text = response.text
    return Text


@app.get("/mocMobChangeCookie")
def mocMobChangeCookie():
    print(request.args)
    token = request.args.get("token")
    print(token)
    params = {
        "token": token,
        "returnUrl": "aHR0cHM6Ly93d3cuaWNvdXJzZTE2My5vcmcv",
    }
    url = "https://www.icourse163.org/passport/logingate/mocMobChangeCookie.htm"

    response = session.get(url=url,params=params)
    Text = response.text
    try:
        setCookieUrlList = re.findall("http.*?setCookie.*?com", Text)
        for i in setCookieUrlList:
            session.get(i)
        session.get("https://www.icourse163.org/")
        cookieDic = dict_from_cookiejar(session.cookies)
        cookieStr = ""
        for key, value in cookieDic.items():
            cookieStr += key + "=" + value + "; "
        session.cookies.clear()
        return cookieStr[:-2]
    except:
        return "登录失败"


if __name__ == '__main__':
    app.run(host="127.0.0.1", port=3001)
# pyinstaller -F loginWeb.py