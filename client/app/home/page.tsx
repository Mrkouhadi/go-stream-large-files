import UploadFile from "../../components/UploadFile";
import DownloadFile from "../../components/DownloadFile";

const Home = () => {
  return (
    <div className="bg-black h-screen">
      <h1>Upload File</h1>
      <UploadFile />
      <hr />
      <h1>Download File</h1>
      <DownloadFile />
    </div>
  );
};

export default Home;
