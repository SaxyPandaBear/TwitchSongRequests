function enableCorsMiddleWare(req, res, next) {
  res.header(
    "Access-Control-Allow-Origin",
    process.env.UI_ENDPOINT || "http://localhost:4200"
  ); // update to match the domain we will make the request from
  res.header(
    "Access-Control-Allow-Headers",
    "Origin, X-Requested-With, Content-Type, Accept"
  );
  res.header("Access-Control-Allow-Methods", "*");
  res.header("Access-Control-Allow-Credentials", "true");
  next();
}
module.exports = enableCorsMiddleWare;
