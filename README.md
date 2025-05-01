#  authkit-react example

An example application demonstrating how to authenticate users with AuthKit's React SDK.
We build on that example by adding a Golang api server where we: 
   1) verify the JWT issues by WorkOS via JWKS and; 
   2) add custom middleware functions and decorators as a foundation for robust access control; leveraging Golang's `context.Context()`. 

## Prerequisites

You will need a [WorkOS account](https://dashboard.workos.com/signup).

## Running the example

1. In the [WorkOS dashboard](https://dashboard.workos.com), head to the Redirects tab and create a [sign-in callback redirect](https://workos.com/docs/user-management/1-configure-your-project/configure-a-redirect-uri) for `http://localhost:5173`.

3. Navigate to the "Authentication" page and click the "Configure CORS" button. Add `http://localhost:5173` to the list of origins to allow clients sessions from. Click "Save."

4. Navigate to the API keys tab and copy the _Client ID_ and _API KEY_.  Rename the `.env.local.example` file to `.env.local` and update your environment variables:

   VITE_WORKOS_CLIENT_ID={_CLIENT ID_}.
   WORKOS_API_KEY={_API KEY_}
5. Start the go server a vscode launch file has been provided in `.vscode/launch.json`
6. Run the following command and navigate to [http://localhost:5173](http://localhost:5173).

   ```bash
   npm run dev
   ```
