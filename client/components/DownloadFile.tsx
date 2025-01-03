"use client";
import { useState } from "react";

const DownloadFile = () => {
  const [downloadProgress, setDownloadProgress] = useState<number>(0);

  const handleDownload=async()=>{
    
      fetch("http://localhost:8080/download?file=Vocabularies.txt", {
        method: "GET",
    })
        .then((response) => {
            if (!response.ok) {
                throw new Error("Failed to download file.");
            }
            return response.blob();
        })
        .then((blob) => {
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = "Vocabularies.txt";
            document.body.appendChild(a);
            a.click();
            a.remove();
            window.URL.revokeObjectURL(url);
        })
        .catch((error) => {
            alert(error.message);
        });    
  }
  // const handleDownload = async () => {
  //   try {
  //     const response = await fetch(
  //       "http://localhost:8080/download?file=Vocabularies.txt",
  //       {
  //         method: "GET",
  //         headers: {
  //           Accept: "application/zip", // Adjust the MIME type according to your file type
  //         },
  //       }
  //     );

  //     if (!response.ok) {
  //       throw new Error(`HTTP error! Status: ${response.status}`);
  //     }

  //     const totalSize = parseInt(
  //       response.headers.get("content-length") || "0",
  //       10
  //     );
  //     const reader = response.body?.getReader();
  //     let receivedSize = 0;

  //     if (!reader) return;

  //     while (true) {
  //       const { done, value } = await reader.read();
  //       if (done) break;

  //       receivedSize += value!.length;
  //       const progress = Math.round((receivedSize / totalSize) * 100);
  //       setDownloadProgress(progress);
  //     }

  //     // Reset download progress after successful download
  //     setDownloadProgress(0);
  //   } catch (error) {
  //     console.error("Error downloading file:", error);
  //     alert("Failed to download file. Please try again.");
  //   }
  // };

  return (
    <div>
      <button onClick={handleDownload}>Download</button>
      {downloadProgress > 0 && <progress value={downloadProgress} max="100" />}
    </div>
  );
};

export default DownloadFile;
