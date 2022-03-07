var result_callback;

function getOnlineNum(cb) {
    result_callback = cb;
    var xhr = new XMLHttpRequest();
    var url = "http://127.0.0.1:9000/getOnLineNum";

    xhr.open("GET", url, true);
    xhr.setRequestHeader("Content-type","application/x-www-form-urlencoded");
    xhr.onreadystatechange = function(){

        var XMLHttpReq = xhr;
        /**
         XMLHttpReq.readyState
         0: 请求未初始化
         1: 服务器连接已建立
         2: 请求已接收
         3: 请求处理中
         4: 请求已完成，且响应已就绪
         **/
        if (XMLHttpReq.readyState == 4) {
            console.log(XMLHttpReq.status);
            if (XMLHttpReq.status == 200) {
                var data = XMLHttpReq.responseText;
                //alert(data)
                console.log(data);
                var json = JSON.parse(data);
                result_callback(0, json)
            }else if(XMLHttpReq.status == 100){

            }else if(XMLHttpReq.status == 300){

            }else if(XMLHttpReq.status == 400){

            }else if(XMLHttpReq.status == 500){

            }else if(XMLHttpReq.status == 0){
                /** 0不是http协议的状态,关于XMLHttpReq.status的说明:
                 1、If the state is UNSENT or OPENED, return 0.（如果状态是UNSENT或者OPENED，返回0）
                 2、If the error flag is set, return 0.（如果错误标签被设置，返回0）
                 3、Return the HTTP status code.（返回HTTP状态码）
                 第一种情况,例如:url请求的是本地文件,状态会是0
                 第二种情况经常出现在跨域请求中,比如url不是本身网站IP或域名,例如请求www.baidu.com时
                 第三种,正常请求本站http协议信息时,正常返回http协议状态值
                 **/
            }

        }
    };
    xhr.send();
}

function getClientList(start, limit, cb) {
    result_callback = cb;
    var xhr = new XMLHttpRequest();
    var url = "http://127.0.0.1:9000/getClientList?start="+start+"&limit="+limit;
    xhr.open("GET", url, true);
    xhr.setRequestHeader("Content-type","application/x-www-form-urlencoded");
    xhr.onreadystatechange = function(){

        var XMLHttpReq = xhr;
        /**
         XMLHttpReq.readyState
         0: 请求未初始化
         1: 服务器连接已建立
         2: 请求已接收
         3: 请求处理中
         4: 请求已完成，且响应已就绪
         **/
        if (XMLHttpReq.readyState == 4) {
            console.log(XMLHttpReq.status);
            if (XMLHttpReq.status == 200) {
                var data = XMLHttpReq.responseText;
                //alert(data)
                console.log(data);
                var json = JSON.parse(data);
                result_callback(1, json)
            }else if(XMLHttpReq.status == 100){

            }else if(XMLHttpReq.status == 300){

            }else if(XMLHttpReq.status == 400){

            }else if(XMLHttpReq.status == 500){

            }else if(XMLHttpReq.status == 0){
                /** 0不是http协议的状态,关于XMLHttpReq.status的说明:
                 1、If the state is UNSENT or OPENED, return 0.（如果状态是UNSENT或者OPENED，返回0）
                 2、If the error flag is set, return 0.（如果错误标签被设置，返回0）
                 3、Return the HTTP status code.（返回HTTP状态码）
                 第一种情况,例如:url请求的是本地文件,状态会是0
                 第二种情况经常出现在跨域请求中,比如url不是本身网站IP或域名,例如请求www.baidu.com时
                 第三种,正常请求本站http协议信息时,正常返回http协议状态值
                 **/
            }

        }
    };
    xhr.send();
}