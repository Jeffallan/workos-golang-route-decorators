// /home/dmitry/projects/pocs/workos-golang-route-decorators/src/routes/api-routes.tsx
import { Box, Flex, Heading, Text, Code, Spinner } from "@radix-ui/themes";
import * as Accordion from '@radix-ui/react-accordion';
import { useUser } from "../hooks/use-user";
import { useAuth } from "@workos-inc/authkit-react";
import { useEffect, useState } from "react";

// Define an interface for the expected response structure from your Go backend
interface ApiResponse {
  message: string;
  extra_context: string;
  claims_message: Record<string, any>;
}

// Helper type for API fetch state
type ApiFetchState = {
  data: ApiResponse | null;
  loading: boolean;
  error: string | null;
};

export default function ApiRoutes() {
  const user = useUser(); // AuthKit user object (might differ from JWT claims)
  const { getAccessToken } = useAuth();

  // State for each API endpoint
  const [rootState, setRootState] = useState<ApiFetchState>({ data: null, loading: true, error: null });
  const [adminState, setAdminState] = useState<ApiFetchState>({ data: null, loading: false, error: null });
  const [userState, setUserState] = useState<ApiFetchState>({ data: null, loading: false, error: null });
  const [orgState, setOrgState] = useState<ApiFetchState>({ data: null, loading: false, error: null });
  const [userZeroState, setUserZeroState] = useState<ApiFetchState>({ data: null, loading: false, error: null });
  const [orgZeroState, setOrgZeroState] = useState<ApiFetchState>({ data: null, loading: false, error: null });


  // State to hold extracted IDs from the root call's claims
  const [userId, setUserId] = useState<string | null>(null);
  const [orgId, setOrgId] = useState<string | null>(null);

  // Generic fetch function
  const fetchDataForEndpoint = async (
    endpoint: string,
    setState: React.Dispatch<React.SetStateAction<ApiFetchState>>
  ): Promise<ApiResponse | null> => {
    setState(prev => ({ ...prev, loading: true, error: null, data: null }));
    let fetchedData: ApiResponse | null = null;

    try {
      const accessToken = await getAccessToken();
      if (!accessToken) {
        throw new Error("Could not retrieve access token.");
      }

      const response = await fetch(`http://localhost:8080${endpoint}`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });

      // Note: We are not checking response.ok here because 401/403 are expected for the ID=0 calls
      // We will rely on the error message parsing below if the status is not 2xx.

      fetchedData = await response.json() as ApiResponse; // Try to parse JSON even for errors

      // If response was not OK, throw an error *after* attempting to parse body
      if (!response.ok) {
          let errorMsg = `HTTP error! Status: ${response.status}`;
          // If JSON parsing worked and there's a message, use it
          if (fetchedData?.message) {
              errorMsg += ` - ${fetchedData.message}`;
          } else {
              // Fallback to text body if JSON parsing failed or no message
              try {
                  const errorBody = await response.text(); // Re-read as text if needed (might fail if already read)
                  errorMsg += ` - ${errorBody}`;
              } catch (_) { /* Ignore if reading body fails */ }
          }
          throw new Error(errorMsg);
      }

      setState(prev => ({ ...prev, data: fetchedData, loading: false }));

    } catch (err: any) {
      console.error(`Error fetching API data for ${endpoint}:`, err);
      // Store the error message, but keep data null
      setState(prev => ({ ...prev, data: null, error: err.message || "An unknown error occurred", loading: false }));
      fetchedData = null; // Ensure we return null on error
    }
    return fetchedData; // Return data for potential chaining/extraction
  };

  // 1. Fetch Root data on mount (and extract IDs)
  useEffect(() => {
    if (!user) {
      setRootState(prev => ({ ...prev, loading: false }));
      return;
    }

    const fetchRootAndExtractIds = async () => {
      const data = await fetchDataForEndpoint("/", setRootState);
      if (data?.claims_message) {
        const claims = data.claims_message;
        // Extract user ID (sub claim)
        if (typeof claims.sub === 'string') {
          setUserId(claims.sub);
        } else {
           console.warn("User ID (sub) not found or not a string in root claims:", claims);
        }
        // Extract org ID (org_id or org claim)
        let extractedOrgId: string | null = null;
        if (typeof claims.org_id === 'string') {
          extractedOrgId = claims.org_id;
        } else if (typeof claims.org === 'string') {
          extractedOrgId = claims.org;
        }
         if (extractedOrgId) {
           setOrgId(extractedOrgId);
         } else {
           console.warn("Org ID (org_id/org) not found or not a string in root claims:", claims);
         }
      }
    };

    fetchRootAndExtractIds();
  }, [user, getAccessToken]); // Depend on user and token getter

  // 2. Fetch Admin data (can run alongside root)
  useEffect(() => {
    if (!user) return; // Only run if logged in
    setAdminState(prev => ({ ...prev, loading: true })); // Set loading true when effect runs
    fetchDataForEndpoint("/admin", setAdminState);
  }, [user, getAccessToken]); // Depend on user and token getter

  // 3. Fetch User data (only when userId is available)
  useEffect(() => {
    if (!user || !userId) return; // Only run if logged in and userId is set
    setUserState(prev => ({ ...prev, loading: true })); // Set loading true when effect runs
    fetchDataForEndpoint(`/user/${userId}`, setUserState);
  }, [userId, user, getAccessToken]); // Depend on userId, user, and token getter

  // 4. Fetch Org data (only when orgId is available)
  useEffect(() => {
    if (!user || !orgId) return; // Only run if logged in and orgId is set
    setOrgState(prev => ({ ...prev, loading: true })); // Set loading true when effect runs
    fetchDataForEndpoint(`/org/${orgId}`, setOrgState);
  }, [orgId, user, getAccessToken]); // Depend on orgId, user, and token getter

  // 5. Fetch User data for ID=0 (run alongside root/admin)
  useEffect(() => {
    if (!user) return;
    setUserZeroState(prev => ({ ...prev, loading: true }));
    fetchDataForEndpoint(`/user/0`, setUserZeroState);
  }, [user, getAccessToken]);

  // 6. Fetch Org data for ID=0 (run alongside root/admin)
  useEffect(() => {
    if (!user) return;
    setOrgZeroState(prev => ({ ...prev, loading: true }));
    fetchDataForEndpoint(`/org/0`, setOrgZeroState);
  }, [user, getAccessToken]);


  // --- User loading/signed-out checks ---
  // Check root loading specifically for the initial page load feel
  if (!user && rootState.loading) {
      return <Spinner size="3" />;
  }
  if (!user && !rootState.loading) {
    return "Please sign in to view API results.";
  }
  // --- End checks ---

  // Generic function to render API call results
  const renderApiResult = (state: ApiFetchState) => {
    if (state.loading) {
      return <Spinner size="2" />;
    }
    // Display error first if it exists
    if (state.error) {
      return <Text color="red">Error: {state.error}</Text>;
    }
    // Display data if it exists (even if there was an error, error takes precedence)
    if (state.data) {
      // Display the fetched data nicely formatted
      return (
        <Code variant="ghost" style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all', display: 'block' }}>
          {JSON.stringify(state.data, null, 2)}
        </Code>
      );
    }
    return <Text color="gray">No data fetched yet or fetch not triggered.</Text>;
  };

  return (
    <>
      <Flex direction="column" gap="2" mb="7">
        <Heading size="8" align="center">
          API Routes
        </Heading>
        <Text size="5" align="center" color="gray">
          Fetching data from protected Go backend routes
        </Text>
      </Flex>

      {/* Accordion container for all API calls */}
      <Flex direction="column" justify="center" gap="3" width="600px"> {/* Wider width */}
        <Accordion.Root type="multiple" collapsible> {/* Allow multiple open */}

          {/* --- GET / Response --- */}
          <Accordion.Item value="get-root-response" mb="3">
            <Accordion.Trigger>
              <Text weight="bold" size="3">GET / Response</Text>
            </Accordion.Trigger>
            <Accordion.Content>
              <Box pt="2">
                {renderApiResult(rootState)}
              </Box>
            </Accordion.Content>
          </Accordion.Item>

          {/* --- GET /admin Response --- */}
          <Accordion.Item value="get-admin-response" mb="3">
            <Accordion.Trigger>
              <Text weight="bold" size="3">GET /admin Response</Text>
            </Accordion.Trigger>
            <Accordion.Content>
              <Box pt="2">
                {renderApiResult(adminState)}
              </Box>
            </Accordion.Content>
          </Accordion.Item>

          {/* --- GET /user/:user_id Response --- */}
          <Accordion.Item value="get-user-response" mb="3">
            <Accordion.Trigger>
              <Text weight="bold" size="3">
                GET /user/{userId ?? '{user_id}'} Response {/* Show placeholder if ID not yet loaded */}
              </Text>
            </Accordion.Trigger>
            <Accordion.Content>
              <Box pt="2">
                {!userId && !rootState.loading && <Text color="gray" size="2">Waiting for User ID from / response...</Text>}
                {userId && renderApiResult(userState)}
              </Box>
            </Accordion.Content>
          </Accordion.Item>

          {/* --- GET /org/:org_id Response --- */}
          <Accordion.Item value="get-org-response" mb="3">
            <Accordion.Trigger>
              <Text weight="bold" size="3">
                GET /org/{orgId ?? '{org_id}'} Response {/* Show placeholder if ID not yet loaded */}
              </Text>
            </Accordion.Trigger>
            <Accordion.Content>
              <Box pt="2">
                 {!orgId && !rootState.loading && <Text color="gray" size="2">Waiting for Org ID from / response...</Text>}
                 {orgId && renderApiResult(orgState)}
              </Box>
            </Accordion.Content>
          </Accordion.Item>

          {/* --- GET /user/0 Response (NEW) --- */}
          <Accordion.Item value="get-user-zero-response" mb="3">
            <Accordion.Trigger>
              {/* Added color="red" */}
              <Text weight="bold" size="3" color="red">
                GET /user/0 Response
              </Text>
            </Accordion.Trigger>
            <Accordion.Content>
              <Box pt="2">
                 {renderApiResult(userZeroState)}
              </Box>
            </Accordion.Content>
          </Accordion.Item>

          {/* --- GET /org/0 Response (NEW) --- */}
          <Accordion.Item value="get-org-zero-response" mb="3">
            <Accordion.Trigger>
              {/* Added color="red" */}
              <Text weight="bold" size="3" color="red">
                GET /org/0 Response
              </Text>
            </Accordion.Trigger>
            <Accordion.Content>
              <Box pt="2">
                 {renderApiResult(orgZeroState)}
              </Box>
            </Accordion.Content>
          </Accordion.Item>

        </Accordion.Root>
      </Flex>
    </>
  );
}
