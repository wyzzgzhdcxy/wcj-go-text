import axios from "axios";

export const w = {}

let toolsUrl = "http://111.229.201.94:39870"
//传递jsonbody对象请求
w.post = function (that, url, body, callback) {
    //如果发起的是请求tools项目的请求，走后端的http代理服务器进行转发
    if (url.startsWith("/tools")){
        url = toolsUrl+url ;
    }
    axios.post(url, body).then(function (ret) {
        let resData = ret.data;
        console.log("请求url:" + url)
        console.dir(resData)
        if (resData.code === 0) {
            callback(resData.data);
        } else {
            that.$message.error("请求异常:" + resData.data)
        }
    }).catch(function (error) {
        console.log(error);
    });
}


w.post1 = function (that, url, body, callback) {
    //如果发起的是请求tools项目的请求，走后端的http代理服务器进行转发
    if (url.startsWith("/tools")){
        url = toolsUrl+url ;
    }
    axios.post(url, body).then(function (ret) {
        let resData = ret.data;
        console.log("请求url:" + url)
        console.dir(resData)
        callback(resData);
    }).catch(function (error) {
        console.log(error);
    });
}


//传递jsonbody对象请求
w.get = function (that, url, params, callback) {
    //如果发起的是请求tools项目的请求，走后端的http代理服务器进行转发
    if (url.startsWith("/tools")){
        url = toolsUrl+url ;
    }
    console.log("请求地址url:" + url)
    axios.get(url, {
        params: params
    }).then(function (response) {
        callback(response.data)
        console.log(response.data);
    }).catch(function (error) {
        that.$message.error("请求异常:" + error)
    });
}


w.postWithParams = function (that, url, params, body, callback) {
    //如果发起的是请求tools项目的请求，走后端的http代理服务器进行转发
    if (url.startsWith("/tools")){
        url = toolsUrl+url ;
    }
    url = url + "?" + new URLSearchParams(params);
    axios.post(url, body).then(function (ret) {
        callback(ret.data)
    }).catch(function (error) {
        console.log(error);
    });
}

w.arrayToSqlQuery = function (values) {
    let sql = "("
    for (let i = 0; i < values.length; i++) {
        if (i !== values.length - 1) {
            sql += "'" + values[i] + "', ";
        } else {
            sql += "'" + values[i] + "'";
        }
    }
    return sql + ")";
}
w.arrayToSQLInsert = function (data) {
    if (Array.isArray(data)) {
        let sql = "";
        for (let i = 0; i < data.length; i++) {
            sql = sql + this.jsonToSQLInsert(data[i]) + "\r\n";
        }
        return sql;
    } else {
        return this.jsonToSQLInsert(data);
    }
}
w.jsonToSQLInsert = function (data) {
    let sql = "INSERT INTO table_name (";
    let keys = Object.keys(data);
    let values = Object.values(data);
    // 构造插入语句的列名部分
    for (let i = 0; i < keys.length; i++) {
        if (i !== keys.length - 1) {
            sql += keys[i] + ", ";
        } else {
            sql += keys[i];
        }
    }
    sql += ") VALUES (";
    // 构造插入语句的值部分
    for (let i = 0; i < values.length; i++) {
        if (i !== values.length - 1) {
            sql += "'" + values[i] + "', ";
        } else {
            sql += "'" + values[i] + "'";
        }
    }
    sql += ")";
    return sql + ";";
}

