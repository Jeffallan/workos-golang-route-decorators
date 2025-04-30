// /home/dmitry/projects/pocs/workos-golang-route-decorators/src/routes/api-routes.tsx
import { Box, Flex, Heading, Text, Code, Spinner } from "@radix-ui/themes"; // Added Code, Spinner
import { useUser } from "../hooks/use-user";
import { useAuth } from "@workos-inc/authkit-react";
import { useEffect, useState } from "react"; // Import useEffect and useState

// Define an interface for the expected response structure from your Go backend
interface ApiResponse {
  message: string;
  extra_context: string;
}

export default function ApiRoutes() {
  const user = useUser();
  const { getAccessToken } = useAuth(); // Destructure getAccessToken

  // State variables to hold the API response, loading status, and errors
  const [apiData, setApiData] = useState<ApiResponse | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true); // Start loading initially
  const [error, setError] = useState<string | null>(null);

  // useEffect to fetch data when the component mounts and user/token are available
  useEffect(() => {
    // Only fetch if we have a user (implies logged in)
    if (!user) {
      setIsLoading(false); // Not logged in, stop loading
      return;
    }

    const fetchData = async () => {
      setIsLoading(true); // Set loading true before fetch
      setError(null); // Clear previous errors
      setApiData(null); // Clear previous data

      try {
        const accessToken = await getAccessToken();
        if (!accessToken) {
          throw new Error("Could not retrieve access token.");
        }

        const response = await fetch("http://localhost:8080/", {
          method: "GET",
          headers: {
            // Include the JWT in the Authorization header
            Authorization: `Bearer ${accessToken}`,
          },
        });

        if (!response.ok) {
          // Try to get error message from backend response body
          let errorMsg = `HTTP error! Status: ${response.status}`;
          try {
            const errorBody = await response.text(); // Or response.json() if backend sends structured errors
            errorMsg += ` - ${errorBody}`;
          } catch (_) {
            // Ignore if reading body fails
          }
          throw new Error(errorMsg);
        }

        const data: ApiResponse = await response.json();
        setApiData(data); // Store the successful response
      } catch (err: any) {
        console.error("Error fetching API data:", err);
        setError(err.message || "An unknown error occurred"); // Store the error message
      } finally {
        setIsLoading(false); // Set loading false after fetch attempt (success or fail)
      }
    };

    fetchData();
  }, [user, getAccessToken]); // Re-run effect if user or getAccessToken changes

  // --- Original user check ---
  if (!user && !isLoading) { // Show loading if user is null but we are still fetching token/user info initially
    return "Please sign in to view API results.";
  }
  if (!user && isLoading) {
      return <Spinner size="3" />; // Show spinner while AuthKit loads user initially
  }
  // --- End original user check ---


  // Function to render the API call results
  const renderApiResult = () => {
    if (isLoading) {
      return <Spinner size="3" />; // Show spinner while fetching
    }
    if (error) {
      return <Text color="red">Error: {error}</Text>; // Show error message
    }
    if (apiData) {
      // Display the fetched data nicely formatted
      return (
        <Code variant="ghost" style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
          {JSON.stringify(apiData, null, 2)}
        </Code>
      );
    }
    return <Text color="gray">No data fetched yet.</Text>; // Fallback message
  };

  return (
    <>
      <Flex direction="column" gap="2" mb="7">
        <Heading size="8" align="center">
          API Routes
        </Heading>
        <Text size="5" align="center" color="gray">
          Fetching data from protected Go backend route `/`
        </Text>
      </Flex>

      {/* Section to display the API call result */}
      <Flex direction="column" justify="center" gap="3" width="500px"> {/* Increased width */}
        <Flex asChild align="center" gap="6">
          <label style={{ alignItems: 'flex-start' }}> {/* Align label top */}
            <Text weight="bold" size="3" style={{ width: 150, flexShrink: 0 }}> {/* Adjusted width */}
              GET / Response
            </Text>
            <Box flexGrow="1">
              {/* Render the result using the helper function */}
              {renderApiResult()}
            </Box>
          </label>
        </Flex>
        {/* You can add more API call sections here */}
        {/* Example:
            <Flex asChild align="center" gap="6">
              <label>
                <Text weight="bold" size="3" style={{ width: 100 }}>
                  GET /admin Response
                </Text>
                <Box flexGrow="1">
                  <p>...</p>
                </Box>
              </label>
            </Flex>
           */}
      </Flex>
    </>
  );
}
