import React from "react";
import ReactDOM from "react-dom/client";
import { createBrowserRouter, redirect, RouterProvider } from "react-router-dom";
import Root from "./routes/root";
import "@radix-ui/themes/styles.css";
import Index from "./routes";
import Account from "./routes/account";
import Login from "./routes/login";
import ApiRoutes from "./routes/api-routes";

function routerGuard(){
  console.log("this is a great place to perform conditional checks for frontend :)")
  // personally I like to keep parts of the decoded JWT in state and perform checks i.e
  // store.getState().value.required
  // if value.required return true
  // for best results you may have to refresh the JWT for the most updated permissions
  // 
  if(true){
    return true
  }
  throw redirect("/login")
}

const router = createBrowserRouter([
  {
    path: "/",
    element: <Root />,
    children: [
      { index: true, element: <Index /> },
      {
        path: "/account",
        element: <Account />,
      },
      {
        path: "/login",
        element: <Login />,
      },
      {
        path: "/api-routes",
        element: <ApiRoutes />,
       loader: () => routerGuard()
      },
    ],
  },
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>
);
